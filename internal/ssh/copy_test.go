package ssh

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
)

var stubSFTPClient = &sftp.Client{}

func stubActor() *Actor {
	// set up an actor where the SFTP client is already "created"
	a := &Actor{
		host:            "test",
		sshClient:       nil,
		_sftpClient:     stubSFTPClient,
		_sftpClientErr:  nil,
		_sftpCreateOnce: sync.Once{},
	}
	a._sftpCreateOnce.Do(func() {})
	return a
}

func failActor() *Actor {
	a := stubActor()
	a._sftpClient = nil
	a._sftpClientErr = fmt.Errorf("test sftp client fail")
	return a
}

func TestActor_CreateRemoteDir(t *testing.T) {
	runtimeStatRemote := statRemote
	runtimeMkdirAllRemote := mkdirAllRemote
	defer func() { statRemote = runtimeStatRemote }()
	defer func() { mkdirAllRemote = runtimeMkdirAllRemote }()

	stubFileInfo := &stubFileInfo{isDir: true}

	var stattedPath string
	stubStatRemote := func(c *sftp.Client, dirPath string) (os.FileInfo, error) {
		stattedPath = dirPath
		assert.Equal(t, stubSFTPClient, c)
		return stubFileInfo, nil
	}
	failStatRemote := func(c *sftp.Client, dirPath string) (os.FileInfo, error) {
		stubStatRemote(c, dirPath) // it's still statted even if there is a failure
		return nil, fmt.Errorf("test fail stat")
	}

	var dirMade string
	stubMkdirAllRemote := func(c *sftp.Client, dirPath string) error {
		dirMade = dirPath
		assert.Equal(t, stubSFTPClient, c)
		return nil
	}
	failMkdirAllRemote := func(c *sftp.Client, dirPath string) error {
		return fmt.Errorf("test fail mkdir all") // dir isn't made if failure
	}

	type wants struct {
		stattedPath string
		dirMade     string
		err         bool
	}
	tests := []struct {
		name        string
		createActor func() *Actor
		dirPath     string
		dirIsFile   bool
		stat        func(c *sftp.Client, dirPath string) (os.FileInfo, error)
		mkdirAll    func(c *sftp.Client, dirPath string) error
		wants       wants
	}{
		// stat success means dir already exists
		{"create success", stubActor, "/root", false, stubStatRemote, stubMkdirAllRemote, wants{
			stattedPath: "/root", dirMade: "not made", err: false}},
		// stat fail means dir should be created
		{"stat failure", stubActor, "/notexist", false, failStatRemote, stubMkdirAllRemote, wants{
			stattedPath: "/notexist", dirMade: "/notexist", err: false}},
		{"dir make failure", stubActor, "/dev", false, failStatRemote, failMkdirAllRemote, wants{
			stattedPath: "/dev", dirMade: "not made", err: true}},
		{"stat is file", stubActor, "/dev/null", true, stubStatRemote, stubMkdirAllRemote, wants{
			stattedPath: "/dev/null", dirMade: "not made", err: true}},
		{"sftp conn fail", failActor, "/home", false, stubStatRemote, stubMkdirAllRemote, wants{
			stattedPath: "not statted", dirMade: "not made", err: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stattedPath = "not statted"
			dirMade = "not made"
			stubFileInfo.isDir = !tt.dirIsFile
			statRemote = tt.stat
			mkdirAllRemote = tt.mkdirAll
			if err := tt.createActor().CreateRemoteDir(tt.dirPath); (err != nil) != tt.wants.err {
				t.Errorf("Actor.CreateRemoteDir() error = %v, want err %v", err, tt.wants.err)
			}
			assert.Equal(t, tt.wants.stattedPath, stattedPath)
			assert.Equal(t, tt.wants.dirMade, dirMade)
		})
	}
}

type stubFileInfo struct {
	isDir bool
}

func (fi *stubFileInfo) Name() string      { return "stubName" }
func (fi *stubFileInfo) Size() int64       { return 1024 }
func (fi *stubFileInfo) Mode() os.FileMode { return 0644 }
func (fi *stubFileInfo) ModTime() time.Time {
	utc, _ := time.LoadLocation("UTC")
	return time.Date(2000, 1, 1, 1, 1, 1, 1, utc)
}
func (fi *stubFileInfo) IsDir() bool      { return fi.isDir }
func (fi *stubFileInfo) Sys() interface{} { return nil }

func TestActor_CopyFileToRemote(t *testing.T) {
	runtimeCreateRemote := createRemote
	runtimeWriteToRemote := writeToRemote
	runtimeCloseRemoteFile := closeRemoteFile
	defer func() { createRemote = runtimeCreateRemote }()
	defer func() { writeToRemote = runtimeWriteToRemote }()
	defer func() { closeRemoteFile = runtimeCloseRemoteFile }()

	stubSourceFile := &os.File{}
	stubSFTPFile := &sftp.File{}

	var fileCreated string
	stubCreateRemote := func(c *sftp.Client, filePath string) (*sftp.File, error) {
		fileCreated = filePath
		assert.Equal(t, stubSFTPClient, c)
		return stubSFTPFile, nil
	}
	failCreateRemote := func(c *sftp.Client, filePath string) (*sftp.File, error) {
		return nil, fmt.Errorf("test fail create")
	}

	stubWriteToRemote := func(dest *sftp.File, source *os.File) (int64, error) {
		assert.Equal(t, stubSFTPFile, dest)
		assert.Equal(t, stubSourceFile, source)
		return 1024, nil
	}
	failWriteToRemote := func(dest *sftp.File, source *os.File) (int64, error) {
		stubWriteToRemote(dest, source)
		return 0, fmt.Errorf("test fail write")
	}

	var remoteFileClosed bool
	closeRemoteFile = func(f *sftp.File) error {
		remoteFileClosed = true
		return nil
	}

	tests := []struct {
		name                 string
		createActor          func() *Actor
		create               func(c *sftp.Client, filePath string) (*sftp.File, error)
		write                func(dest *sftp.File, source *os.File) (int64, error)
		remoteFilePath       string
		fileCreated          string
		wantRemoteFileClosed bool
		wantErr              bool
	}{
		{"copy success", stubActor, stubCreateRemote, stubWriteToRemote, "/etc/wasp",
			"/etc/wasp", true, false},
		{"create fail", stubActor, failCreateRemote, stubWriteToRemote, "/dev/devices",
			"not created", false, true},
		{"write fail", stubActor, stubCreateRemote, failWriteToRemote, "/mnt/unmounted",
			"/mnt/unmounted", true, true},
		{"sftp conn fail", failActor, stubCreateRemote, stubWriteToRemote, "/home/drake",
			"not created", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileCreated = "not created"
			remoteFileClosed = false
			createRemote = tt.create
			writeToRemote = tt.write
			if err := tt.createActor().CopyFileToRemote(stubSourceFile, tt.remoteFilePath); (err != nil) != tt.wantErr {
				t.Errorf("Actor.CopyFileToRemote() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.fileCreated, fileCreated)
			assert.Equal(t, tt.wantRemoteFileClosed, remoteFileClosed)
		})
	}
}
