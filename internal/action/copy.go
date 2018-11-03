package action

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/util"
	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
)

const (
	// max file pointers defaults to 1024 on my system; set a reasonable default here that won't
	// overwhelm the system.
	maxFilePointers = 512
)

// Counting semaphore to limit the number of files Octopus will open on the host.
// The default file limit is 1024. Don't stress the system too much.
var filePointers = make(chan struct{}, maxFilePointers)

// FileCopier is a tentacle action which copies local files to a remote host.
type FileCopier struct {
	LocalSources []string
	RemoteDir    string
	Recursive    bool
}

// Do executes the command tentacle's command on the remote host.
func (c *FileCopier) Do(host string, client *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	logger.Info.Println("establishing sftp connection to host:", host)
	sftp, err := sftp.NewClient(client)
	if err != nil {
		err = fmt.Errorf("unable to start sftp subsystem for host %s: %+v", host, err)
		return
	}
	defer sftp.Close()

	if err = createRemoteDir(sftp, c.RemoteDir); err != nil {
		return
	}

	errCh := make(chan error, maxFilePointers)
	var wg sync.WaitGroup

	for _, s := range c.LocalSources {
		var fp string
		fp, err = util.AbsPath(s)
		if err != nil {
			err = fmt.Errorf("cannot start copying files: %+v", err)
			return
		}
		wg.Add(1)
		go doCopyDirOrFile(sftp, fp, c.RemoteDir, c.Recursive, &wg, errCh)
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

func doCopyDirOrFile(
	client *sftp.Client,
	sourcePath, destDir string,
	recursive bool,
	wg *sync.WaitGroup, errors chan<- error,
) {
	defer wg.Done()

	if fi, err := os.Stat(sourcePath); err != nil {
		errors <- fmt.Errorf("could not get info about source path %s: %+v", sourcePath, err)
		return
	} else if fi.IsDir() && !recursive {
		errors <- fmt.Errorf("skipping local path %s because it is a directory and recursive copy is not enabled", sourcePath)
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
			if err := createRemoteDir(client, fullDest); err != nil {
				errors <- err
				// don't double report this err in 'error walking local dir ...'
				return filepath.SkipDir
			}
		} else {
			wg.Add(1)
			go doCopyFile(client, pth, fullDest, wg, errors)
		}

		return nil
	})
	if err != nil {
		errors <- fmt.Errorf("error walking local dir %s: %+v", sourcePath, err)
	}
}

func doCopyFile(
	client *sftp.Client,
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

	d, err := client.Create(destPath)
	if err != nil {
		errors <- fmt.Errorf("could not open remote file %s for writing: %+v", destPath, err)
		return
	}
	defer d.Close()

	if _, err := d.WriteTo(s); err != nil {
		errors <- fmt.Errorf("failed to copy file %s to remote at %s: %+v", sourcePath, destPath, err)
		return
	}
}

func createRemoteDir(client *sftp.Client, dir string) error {
	if fi, err := client.Stat(dir); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("specified remote dir %s is a file, not a dir; cannot copy files", dir)
		}
		// already exists
	} else if err := client.MkdirAll(dir); err != nil {
		return fmt.Errorf("could not create remote directory %s: %+v", dir, err)
	}
	return nil
}

func min(vals ...int) int {
	var min int
	for i, v := range vals {
		if i == 0 {
			min = v
		} else if v < min {
			min = v
		}
	}
	return min
}
