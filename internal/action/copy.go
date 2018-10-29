package action

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BlaineEXE/octopus/internal/logger"

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

// CopyFiles is a tentacle action which copies local files to a remote host.
type CopyFiles struct {
	LocalSources []string
	RemoteDir    string
}

// Do executes the command tentacle's command on the remote host.
func (c *CopyFiles) Do(context *Context) (*Data, error) {
	data := &Data{
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	}

	logger.Info.Println("establishing sftp connection to host:", context.Host)
	sftp, err := sftp.NewClient(context.Client)
	if err != nil {
		return nil, fmt.Errorf("unable to start sftp subsystem for host %s: %+v", context.Host, err)
	}
	defer sftp.Close()

	resCh := make(chan copyResult, maxFilePointers)

	go func() {
		for _, s := range c.LocalSources {
			go func(s string) {
				err := doCopyFile(sftp, s, c.RemoteDir)
				resCh <- copyResult{s, err}
			}(s)
		}
	}()

	numFail := 0
	for range c.LocalSources {
		r := <-resCh
		if r.err != nil {
			numFail++
			// append fail message to stderr
			data.Stderr.WriteString(fmt.Sprintf("%+v\n", r.err))
		}
	}

	data.Stdout.WriteString(fmt.Sprintf("wrote %d of %d files",
		len(c.LocalSources)-numFail, len(c.LocalSources)))
	e := error(nil)
	if numFail > 0 {
		e = fmt.Errorf("failed to copy %d out of %d files to host %s",
			numFail, len(c.LocalSources), context.Host)
	}
	return data, e
}

type copyResult struct {
	localSource string
	err         error
}

func doCopyFile(client *sftp.Client, sourcePath, destDir string) error {
	filePointers <- struct{}{} // claim a file pointer resource
	s, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("could not open local file %s for reading: %+v", sourcePath, err)
	}
	defer func() {
		s.Close()
		<-filePointers // release a file pointer resource
	}()

	// Creating dir can panic if using a dir that already exists (e.g., /test, /dev/null), so
	// check if it already exists to avoid this
	if err, _ := client.ReadDir(destDir); err == nil {
		// Dir already exists
	} else if err := client.MkdirAll(destDir); err != nil {
		return fmt.Errorf("could not create remote directory %s: %+v", destDir, err)
	}

	filename := filepath.Base(sourcePath)
	destPath := filepath.Join(destDir, filename)
	d, err := client.Create(destPath)
	if err != nil {
		return fmt.Errorf("could not open remote file %s for writing: %+v", destPath, err)
	}
	defer d.Close()

	if _, err := d.WriteTo(s); err != nil {
		return fmt.Errorf("failed to copy file %s to remote at %s: %+v", sourcePath, destPath, err)
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
