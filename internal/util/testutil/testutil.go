package testutil

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
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
}

// CompareStringLists ensures that the expected and got string lists have exactly the same elements.
func CompareStringLists(t *testing.T, expected []string, got []string, name string) {
	g := make([]string, len(got))
	copy(g, got)
	var err error
	for _, e := range expected {
		g, err = removeFromList(e, g)
		//fmt.Println(g)
		if err != nil {
			t.Errorf("%s: %+v", name, err)
		}
	}
	if len(g) > 0 {
		t.Errorf("%s has extraneous items: %+v", name, g)
	}
}

func removeFromList(key string, list []string) ([]string, error) {
	for i, s := range list {
		if s == key {
			return append(list[:i], list[i+1:]...), nil // remove element & return new list
		}
	}
	return list, fmt.Errorf("%s not in list %+v", key, list)
}
