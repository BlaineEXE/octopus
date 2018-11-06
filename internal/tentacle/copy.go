package tentacle

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BlaineEXE/octopus/internal/remote"
	"github.com/BlaineEXE/octopus/internal/util"
)

const (
	// max file pointers defaults to 1024 on my system; set a reasonable default here that won't
	// overwhelm the system.
	maxFilePointers = 512
)

// Counting semaphore to limit the number of files Octopus will open on the local host.
// The default file limit is 1024. Don't stress the system too much.
var filePointers = make(chan struct{}, maxFilePointers)

// CopyFileOptions is a collection of additional options for how files are copied to remote hosts.
type CopyFileOptions struct {
	recursive bool
	// TODO: recursive is the first option, but many more may follow. E.g., follow-symlinks,
	// copy-symlinks, preserve-attributes, ...
}

// NewCopyFileOptions creates a new option struct for defining how files are to be copied.
// Create new options in a function instead of relying on a struct so developers are less likely to
// leave a newly created option unset.
func NewCopyFileOptions(recursive bool) *CopyFileOptions {
	return &CopyFileOptions{
		recursive: recursive,
	}
}

// FileCopier returns a new remote action definition which defines how local files/dirs are to be
// copied to an actor's remote host.
func FileCopier(
	localSourcePaths []string,
	remoteDestDir string,
	opts *CopyFileOptions,
) remote.Action {
	return func(a remote.Actor) (stdout, stderr *bytes.Buffer, err error) {
		if err = a.CreateRemoteDir(remoteDestDir); err != nil {
			return
		}

		errCh := make(chan error, maxFilePointers)
		var wg sync.WaitGroup

		for _, s := range localSourcePaths {
			var fp string
			fp, err = util.AbsPath(s)
			if err != nil {
				err = fmt.Errorf("cannot start copying files: %+v", err)
				return
			}
			wg.Add(1)
			go doCopyDirOrFile(a, fp, remoteDestDir, opts.recursive, &wg, errCh)
		}

		// Close the channel when all files are copied (which could be recursively)
		go func() {
			wg.Wait()
			close(errCh)
		}()

		stdout = new(bytes.Buffer)
		stderr = new(bytes.Buffer)

		numFail := 0
		for err := range errCh {
			if err != nil {
				numFail++
				// append fail message to stderr
				stderr.WriteString(fmt.Sprintf("%+v\n", err))
			}
		}

		err = error(nil)
		if numFail > 0 {
			err = fmt.Errorf("failed to copy %d path(s)", numFail)
		} else {
			stdout.WriteString("wrote all files")
		}
		return
	}
}

// if it's a dir, walk the tree and copy each file; if it's a file, just copy it
func doCopyDirOrFile(
	a remote.Actor,
	sourcePath, destDir string,
	recursive bool,
	wg *sync.WaitGroup, errors chan<- error,
) {
	defer wg.Done()

	if fi, err := os.Stat(sourcePath); err != nil {
		errors <- fmt.Errorf("could not get info about source path %s: %+v", sourcePath, err)
		return
	} else if fi.IsDir() && !recursive {
		errors <- fmt.Errorf(
			"skipping local path %s because it is a directory and recursive copy is not enabled", sourcePath)
		return
	}

	// This works for a single file or for a dir
	sourceRoot, _ := filepath.Split(sourcePath)
	err := filepath.Walk(sourcePath, func(pth string, info os.FileInfo, err error) error {
		if err != nil {
			errors <- fmt.Errorf("could not access local dir or file %s: %+v", pth, err)
			// don't double report this err in 'error walking local dir ...'
			return filepath.SkipDir
		}

		relPath := pth[len(sourceRoot):]
		fullDest := filepath.Join(destDir, relPath)
		if info.IsDir() {
			// Source base is a dir, and we want to include this base dir on the host.
			if err := a.CreateRemoteDir(fullDest); err != nil {
				errors <- err
				// don't double report this err in 'error walking local dir ...'
				return filepath.SkipDir
			}
		} else {
			wg.Add(1)
			go doCopyFile(a, pth, fullDest, wg, errors)
		}

		return nil
	})
	if err != nil {
		errors <- fmt.Errorf("error walking local dir %s: %+v", sourcePath, err)
	}
}

// copy a single file to remote
func doCopyFile(
	a remote.Actor,
	sourcePath, destPath string,
	wg *sync.WaitGroup, errors chan<- error,
) {
	defer wg.Done()

	filePointers <- struct{}{}        // claim a file pointer resource
	defer func() { <-filePointers }() // release a file pointer resource on any return
	s, err := os.Open(sourcePath)
	if err != nil {
		errors <- fmt.Errorf("could not open local file %s for reading: %+v", sourcePath, err)
		return
	}
	defer s.Close()

	if err := a.CopyFileToRemote(s, destPath); err != nil {
		errors <- fmt.Errorf("failed to copy file %s to remote at %s. %+v", sourcePath, destPath, err)
		return
	}
}
