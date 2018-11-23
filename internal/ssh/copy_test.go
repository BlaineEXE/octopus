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

func mockActor(failSFTPConn bool) *Actor {
	// set up an actor where the SFTP client is already "created"
	a := &Actor{
		host:            "test",
		sshClient:       nil,
		_sftpClient:     stubSFTPClient,
		_sftpClientErr:  nil,
		_sftpCreateOnce: sync.Once{},
	}
	a._sftpCreateOnce.Do(func() {})
	if failSFTPConn {
		a._sftpClient = nil
		a._sftpClientErr = fmt.Errorf("test sftp client fail")
	}
	return a
}

func TestActor_CreateRemoteDir(t *testing.T) {
	runtimeStatRemote := statRemote
	runtimeMkdirAllRemote := mkdirAllRemote
	defer func() { statRemote = runtimeStatRemote }()
	defer func() { mkdirAllRemote = runtimeMkdirAllRemote }()

	stubFileInfo := &mockFileInfo{isDir: true}

	var failStatRemote bool
	var stattedPath string
	statRemote = func(c *sftp.Client, dirPath string) (os.FileInfo, error) {
		stattedPath = dirPath // dir is still statted even if there is a failure
		assert.Equal(t, stubSFTPClient, c)
		if failStatRemote {
			return nil, fmt.Errorf("test fail stat")
		}
		return stubFileInfo, nil
	}

	var failMkdirAll bool
	var dirMade string
	var modeMade os.FileMode
	mkdirAllRemote = func(c *sftp.Client, dirPath string, mode os.FileMode) error {
		assert.Equal(t, stubSFTPClient, c)
		if failMkdirAll {
			return fmt.Errorf("test fail mkdir all") // dir isn't made if failure
		}
		dirMade = dirPath
		modeMade = mode
		return nil
	}

	type args struct {
		dirPath string
		mode    os.FileMode
	}
	type inject struct {
		dirIsFile    bool
		failStat     bool
		failMkdirAll bool
		failSFTPConn bool
	}
	type wants struct {
		stattedPath string
		err         bool
	}
	tests := []struct {
		name   string
		args   args
		inject inject
		wants  wants
	}{
		// stat success means dir already exists
		{"create success", args{"/root", 0644}, inject{},
			wants{stattedPath: "/root", err: false}},
		// stat fail means dir should be created
		{"stat failure", args{"/notexist", 0755}, inject{failStat: true},
			wants{stattedPath: "/notexist", err: false}},
		{"dir make failure", args{"/dev", 0750}, inject{failStat: true, failMkdirAll: true},
			wants{stattedPath: "/dev", err: true}},
		{"stat is file", args{"/dev/null", 0500}, inject{dirIsFile: true},
			wants{stattedPath: "/dev/null", err: true}},
		{"sftp conn fail", args{"/home", 0600}, inject{failSFTPConn: true},
			wants{stattedPath: "not statted", err: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stattedPath = "not statted"
			dirMade = "not made"
			modeMade = os.FileMode(0000)
			stubFileInfo.isDir = !tt.inject.dirIsFile
			failStatRemote = tt.inject.failStat
			failMkdirAll = tt.inject.failMkdirAll
			err := mockActor(tt.inject.failSFTPConn).CreateRemoteDir(tt.args.dirPath, tt.args.mode)
			if (err != nil) != tt.wants.err {
				t.Errorf("Actor.CreateRemoteDir() error = %v, want err %v", err, tt.wants.err)
			}
			// dir always statted unless error creating actor
			if !tt.inject.failSFTPConn {
				assert.Equal(t, tt.args.dirPath, stattedPath)
			} else {
				assert.Equal(t, "not statted", stattedPath)
			}
			// dir isn't made if error creating actor, failure to make dir, if dir already exists
			if !tt.inject.failSFTPConn && !failMkdirAll && tt.inject.failStat && !tt.inject.dirIsFile {
				assert.Equal(t, tt.args.dirPath, dirMade)
				assert.Equal(t, tt.args.mode, modeMade)
			} else {
				assert.Equal(t, "not made", dirMade)
				assert.Equal(t, os.FileMode(0000), modeMade)
			}
		})
	}
}

type mockFileInfo struct {
	isDir bool
}

func (fi *mockFileInfo) Name() string      { return "stubName" }
func (fi *mockFileInfo) Size() int64       { return 1024 }
func (fi *mockFileInfo) Mode() os.FileMode { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time {
	utc, _ := time.LoadLocation("UTC")
	return time.Date(2000, 1, 1, 1, 1, 1, 1, utc)
}
func (fi *mockFileInfo) IsDir() bool      { return fi.isDir }
func (fi *mockFileInfo) Sys() interface{} { return nil }

func TestActor_CopyFileToRemote(t *testing.T) {
	runtimeCreateRemote := createRemote
	runtimeWriteToRemote := writeToRemote
	runtimeCloseRemoteFile := closeRemoteFile
	defer func() { createRemote = runtimeCreateRemote }()
	defer func() { writeToRemote = runtimeWriteToRemote }()
	defer func() { closeRemoteFile = runtimeCloseRemoteFile }()

	stubSourceFile := &os.File{}
	stubSFTPFile := &sftp.File{}

	var failCreateRemote bool
	var fileCreated string
	var modeCreated os.FileMode
	createRemote = func(c *sftp.Client, filePath string, mode os.FileMode) (*sftp.File, error) {
		assert.Equal(t, stubSFTPClient, c)
		if failCreateRemote {
			return nil, fmt.Errorf("test fail create")
		}
		fileCreated = filePath
		modeCreated = mode
		return stubSFTPFile, nil
	}

	var failWriteToRemote bool
	writeToRemote = func(dest *sftp.File, source *os.File) (int64, error) {
		assert.Equal(t, stubSFTPFile, dest)
		assert.Equal(t, stubSourceFile, source)
		if failWriteToRemote {
			return 0, fmt.Errorf("test fail write")
		}
		return 1024, nil
	}

	var remoteFileClosed bool
	closeRemoteFile = func(f *sftp.File) error {
		remoteFileClosed = true
		return nil
	}

	type args struct {
		remoteFilePath string
		mode           os.FileMode
	}
	type inject struct {
		failSFTPConn      bool
		failCreateRemote  bool
		failWriteToRemote bool
	}
	type wants struct {
		err bool
	}
	tests := []struct {
		name   string
		args   args
		inject inject
		wants  wants
	}{
		{"copy success", args{"/etc/wasp", 0644}, inject{},
			wants{err: false}},
		{"create fail", args{"/dev/devices", 0755}, inject{failCreateRemote: true},
			wants{err: true}},
		{"write fail", args{"/mnt/unmounted", 0750}, inject{failWriteToRemote: true},
			wants{err: true}},
		{"sftp conn fail", args{"/home/drake", 0500}, inject{failSFTPConn: true},
			wants{err: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileCreated = "not created"
			modeCreated = os.FileMode(0000)
			remoteFileClosed = false
			failCreateRemote = tt.inject.failCreateRemote
			failWriteToRemote = tt.inject.failWriteToRemote
			err := mockActor(tt.inject.failSFTPConn).
				CopyFileToRemote(stubSourceFile, tt.args.remoteFilePath, tt.args.mode)
			if (err != nil) != tt.wants.err {
				t.Errorf("Actor.CopyFileToRemote() error = %v, wantErr %v", err, tt.wants.err)
			}
			// file is only copied if create file doesn't fail and create actor doesn't fail
			if tt.inject.failSFTPConn || failCreateRemote {
				assert.Equal(t, "not created", fileCreated)
				assert.Equal(t, os.FileMode(0000), modeCreated)
				assert.Equal(t, false, remoteFileClosed)
			} else {
				assert.Equal(t, tt.args.remoteFilePath, fileCreated)
				assert.Equal(t, tt.args.mode, modeCreated)
				assert.Equal(t, true, remoteFileClosed)
			}
		})
	}
}
