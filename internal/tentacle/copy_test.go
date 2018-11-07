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
	defer os.RemoveAll(tmpRoot)
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
	for _, f := range allFiles {
		os.MkdirAll(path.Dir(f), 0777)
		testutil.WriteFile(f, path.Base(f), 0777) // file's text is its filename
	}
	os.Chmod(fileWO, 0222)
	os.Chmod(dirWO, 0222)

	recursive := NewCopyFileOptions(true)
	notRecursive := NewCopyFileOptions(false)

	type args struct {
		localSourcePaths []string
		remoteDestDir    string
		opts             *CopyFileOptions
	}
	type wants struct {
		dirs      []string
		dirFails  []string
		files     []string
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
				files: []string{"/rmt/fileA", "/rmt/fileAB", "/rmt/fileBA"},
				dirs:  []string{"/rmt"},
				err:   false}},
		{"copy files but not dir to remote when recursive is false",
			args{[]string{fileA, dirA, fileBA}, "/home", notRecursive},
			&remotetest.MockRemoteActor{},
			wants{
				files: []string{"/home/fileA", "/home/fileBA"},
				dirs:  []string{"/home"},
				err:   true}},
		{"cannot create remote root dir",
			args{[]string{fileA, fileB}, "/nope", notRecursive},
			&remotetest.MockRemoteActor{CreateDirErrorOn: "/nope"},
			wants{
				dirs:     []string{"/nope"},
				dirFails: []string{"/nope"},
				err:      true}},
		{"copy dirs and files when recursive is true",
			args{[]string{fileBA, dirA}, "/etc", recursive},
			&remotetest.MockRemoteActor{},
			wants{
				// since fileBA is specified directly, it should be copied into the remote root
				files: []string{"/etc/fileBA", "/etc/dirA/fileAA", "/etc/dirA/fileAB"},
				dirs:  []string{"/etc", "/etc/dirA"},
				err:   false}},
		{"fail to copy write-only dirs and files",
			args{[]string{fileWO, dirWO, fileWOA}, "/root", recursive},
			&remotetest.MockRemoteActor{},
			wants{
				dirs: []string{"/root"},
				err:  true}},
		{"cannot create remote dir",
			args{[]string{dirA, fileB}, "/tmp", recursive},
			&remotetest.MockRemoteActor{CreateDirErrorOn: "dirA"},
			wants{
				dirs:     []string{"/tmp", "/tmp/dirA"},
				dirFails: []string{"/tmp/dirA"},
				files:    []string{"/tmp/fileB"}, // do not attempt to copy files in tmp
				err:      true}},
		{"cannot create remote file",
			args{[]string{dirA, fileB}, "/dev", recursive},
			&remotetest.MockRemoteActor{CopyFileErrorOn: "fileAA"},
			wants{
				dirs:      []string{"/dev", "/dev/dirA"},
				files:     []string{"/dev/fileB", "/dev/dirA/fileAA", "/dev/dirA/fileAB"},
				fileFails: []string{"/dev/dirA/fileAA"},
				err:       true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := tt.actor
			action := FileCopier(tt.args.localSourcePaths, tt.args.remoteDestDir, tt.args.opts)
			_, e, err := action(a)
			fmt.Println(e)
			fmt.Println(err)
			assert.True(t, (err != nil) == tt.wants.err) //err received when expected

			assert.Zero(t, len(a.Commands))
			dirs := remotetest.Clear(&a.DirCreates)
			dirFails := remotetest.Clear(&a.DirCreateFails)
			files := remotetest.Clear(&a.FileCopies)
			fileFails := remotetest.Clear(&a.FileCopyFails)
			testutil.CompareStringLists(t, tt.wants.dirs, dirs, "DirCreates")
			testutil.CompareStringLists(t, tt.wants.dirFails, dirFails, "DirCreateFails")
			testutil.CompareStringLists(t, tt.wants.files, files, "FileCopies")
			testutil.CompareStringLists(t, tt.wants.fileFails, fileFails, "FileCopyFails")
		})
	}
}

// func compareLists(t *testing.T, expected []string, got []string, name string) {
// 	g := make([]string, len(got))
// 	copy(g, got)
// 	var err error
// 	for _, e := range expected {
// 		g, err = removeFromList(e, g)
// 		//fmt.Println(g)
// 		if err != nil {
// 			t.Errorf("%s: %+v", name, err)
// 		}
// 	}
// 	if len(g) > 0 {
// 		t.Errorf("%s has extraneous items: %+v", name, g)
// 	}
// }

// func removeFromList(key string, list []string) ([]string, error) {
// 	for i, s := range list {
// 		if s == key {
// 			return append(list[:i], list[i+1:]...), nil // remove element & return new list
// 		}
// 	}
// 	return list, fmt.Errorf("%s not in list %+v", key, list)
// }
