// Package util provides miscellaneous utility functions.
// This includes path functions for processing filesystem paths.
package util

import (
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
)

// AbsPath returns the absolute path representing a given path with environment variables in the
// path expanded as well as special homedir symbols expanded (e.g. `~`).
func AbsPath(path string) (string, error) {
	p, err := homedir.Expand(path) // can expand '~' but can't expand '$HOME'
	if err != nil {
		return "", fmt.Errorf("error parsing path %s: %+v", path, err)
	}
	p = os.ExpandEnv(p) // can expand '$HOME' but cannot expand '~'
	a, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for %s: %+v", path, err)
	}
	return a, nil
}
