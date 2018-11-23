package testutil

import (
	"io/ioutil"
	"log"
	"os"
)

// TempDir creates a new temp dir with prefix for testing. Returns dir name and cleanup function.
func TempDir(prefix string) (dirPath string, cleanup func()) {
	dirPath, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("failed to create temp dir for testing. %+v", err)
	}
	return dirPath, func() { os.RemoveAll(dirPath) }
}

// WriteFile writes the contents to the file at path.
func WriteFile(path, contents string, perms os.FileMode) {
	if err := ioutil.WriteFile(path, []byte(contents), perms); err != nil {
		log.Fatalf("failed to create test file %s. %+v", path, err)
	}
	if err := os.Chmod(path, perms); err != nil {
		log.Fatalf("failed to set perms on test file %s. %+v", path, err)
	}
}
