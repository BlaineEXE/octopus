package tentacle

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/BlaineEXE/octopus/internal/util/testutil"

	"github.com/stretchr/testify/assert"

	remotetest "github.com/BlaineEXE/octopus/internal/remote/test"
)

func TestFileCopier(t *testing.T) {
	tmpRoot, cleanup := testutil.TempDir("")
	defer cleanup()
	fileA := path.Join(tmpRoot, "fileA")
	fileB := path.Join(tmpRoot, "fileB")
	dirA := path.Join(tmpRoot, "dirA")
	fileAA := path.Join(dirA, "fileAA")
	fileAB := path.Join(dirA, "fileAB")
	dirB := path.Join(tmpRoot, "dirB")
	fileBA := path.Join(dirB, "fileBA")
	fileWO := path.Join(tmpRoot, "fileWO") // write only file
	dirWO := path.Join(tmpRoot, "dirWO")   // write only dir
	fileWOA := path.Join(dirWO, "fileWOA") // rw file in write only dir
	allFiles := []string{fileA, fileB, fileAA, fileAB, fileBA, fileWO, fileWOA}
	allModes := []os.FileMode{0755, 0750, 0700, 0760, 0770, 0222, 0777}
	for i, f := range allFiles {
		os.MkdirAll(path.Dir(f), 0744)                   // must be 07-- or test will fail
		testutil.WriteFile(f, path.Base(f), allModes[i]) // file's text is its filename
	}
	os.Chmod(dirB, 0777)
	os.Chmod(dirWO, 0222)
	defer os.Chmod(dirWO, 0777) // need to be able to delete this later

	recursive := NewCopyFileOptions(true)
	notRecursive := NewCopyFileOptions(false)

	type args struct {
		localSourcePaths []string
		remoteDestDir    string
		opts             *CopyFileOptions
	}
	type wants struct {
		dirs      []string
		dirModes  []os.FileMode
		dirFails  []string
		files     []string
		fileModes []os.FileMode
		fileFails []string
		err       bool
	}
	tests := []struct {
		name  string
		args  args
		actor *remotetest.MockRemoteActor
		wants wants
	}{
		{"copy files to remote dir",
			args{[]string{fileA, fileAB, fileBA}, "/rmt", notRecursive},
			&remotetest.MockRemoteActor{},
			wants{
				files:     []string{"/rmt/fileA", "/rmt/fileAB", "/rmt/fileBA"},
				fileModes: []os.FileMode{0755, 0760, 0770},
				dirs:      []string{"/rmt"},
				dirModes:  []os.FileMode{0644}, // base dir always created with 0644
				err:       false}},
		{"copy files but not dir to remote when recursive is false",
			args{[]string{fileA, dirA, fileBA}, "/home", notRecursive},
			&remotetest.MockRemoteActor{},
			wants{
				files:     []string{"/home/fileA", "/home/fileBA"},
				fileModes: []os.FileMode{0755, 0770},
				dirs:      []string{"/home"},
				dirModes:  []os.FileMode{0644},
				err:       true}},
		{"cannot create remote root dir",
			args{[]string{fileA, fileB}, "/nope", notRecursive},
			&remotetest.MockRemoteActor{CreateDirErrorOn: "/nope"},
			wants{
				dirs:     []string{"/nope"},
				dirModes: []os.FileMode{0644},
				dirFails: []string{"/nope"},
				err:      true}},
		{"copy dirs and files when recursive is true",
			args{[]string{fileBA, dirA}, "/etc", recursive},
			&remotetest.MockRemoteActor{},
			wants{
				// since fileBA is specified directly, it should be copied into the remote root
				files:     []string{"/etc/fileBA", "/etc/dirA/fileAA", "/etc/dirA/fileAB"},
				fileModes: []os.FileMode{0770, 0700, 0760},
				dirs:      []string{"/etc", "/etc/dirA"},
				dirModes:  []os.FileMode{0644, 0744},
				err:       false}},
		{"fail to copy write-only dirs and files",
			args{[]string{fileWO, dirWO, fileWOA}, "/root", recursive},
			&remotetest.MockRemoteActor{},
			wants{
				dirs:     []string{"/root"},
				dirModes: []os.FileMode{0644},
				err:      true}},
		{"cannot create remote dir",
			args{[]string{dirA, fileB}, "/tmp", recursive},
			&remotetest.MockRemoteActor{CreateDirErrorOn: "dirA"},
			wants{
				dirs:      []string{"/tmp", "/tmp/dirA"},
				dirModes:  []os.FileMode{0644, 0744},
				dirFails:  []string{"/tmp/dirA"},
				files:     []string{"/tmp/fileB"}, // do not attempt to copy files in tmp
				fileModes: []os.FileMode{0750},
				err:       true}},
		{"cannot create remote file",
			args{[]string{dirA, fileB}, "/dev", recursive},
			&remotetest.MockRemoteActor{CopyFileErrorOn: "fileAA"},
			wants{
				dirs:      []string{"/dev", "/dev/dirA"},
				dirModes:  []os.FileMode{0644, 0744},
				files:     []string{"/dev/fileB", "/dev/dirA/fileAA", "/dev/dirA/fileAB"},
				fileModes: []os.FileMode{0750, 0700, 0760},
				fileFails: []string{"/dev/dirA/fileAA"},
				err:       true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.actor
			a.DirCreates = []string{}
			a.DirCreateModes = []os.FileMode{}
			a.DirCreateFails = []string{}
			a.FileCopies = []string{}
			a.FileCopyModes = []os.FileMode{}
			a.FileCopyFails = []string{}
			action := FileCopier(tt.args.localSourcePaths, tt.args.remoteDestDir, tt.args.opts)
			_, e, err := action(a)
			fmt.Println(e)
			fmt.Println(err)
			assert.True(t, (err != nil) == tt.wants.err) //err received when expected

			assert.Zero(t, len(a.Commands))
			assert.ElementsMatch(t, tt.wants.dirs, a.DirCreates, "DirCreates")
			assert.ElementsMatch(t, tt.wants.dirModes, a.DirCreateModes, "DirCreateModes")
			assert.ElementsMatch(t, tt.wants.dirFails, a.DirCreateFails, "DirCreateFails")
			assert.ElementsMatch(t, tt.wants.files, a.FileCopies, "FileCopies")
			assert.ElementsMatch(t, tt.wants.fileModes, a.FileCopyModes, "FileCopyModes")
			assert.ElementsMatch(t, tt.wants.fileFails, a.FileCopyFails, "FileCopyFails")
		})
	}
}
