package ssh

import (
	"fmt"
	"os"

	"github.com/pkg/sftp"
)

var statRemote = func(c *sftp.Client, dirPath string) (os.FileInfo, error) {
	return c.Stat(dirPath)
}

var mkdirAllRemote = func(c *sftp.Client, dirPath string) error {
	return c.MkdirAll(dirPath)
}

// CreateRemoteDir creates the dir as well as any nonexistent parents on the Actor's remote host if
// any of the dirs do not exist. Return nil if the paths already exist.
func (a *Actor) CreateRemoteDir(dirPath string) error {
	errMsg := "failed to create remote dir " + dirPath + ". %+v"
	c, err := a.sftpClient()
	if err != nil {
		return err
	}
	if fi, err := statRemote(c, dirPath); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf(errMsg, "dir exists and is a file")
		}
		// already exists
	} else if err := mkdirAllRemote(c, dirPath); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

var createRemote = func(c *sftp.Client, filePath string) (*sftp.File, error) {
	return c.Create(filePath)
}

var writeToRemote = func(dest *sftp.File, source *os.File) (int64, error) {
	return dest.ReadFrom(source)
}

var closeRemoteFile = func(f *sftp.File) error {
	return f.Close()
}

// CopyFileToRemote copies the file to the Actor's remote host at the remote file path.
func (a *Actor) CopyFileToRemote(localSource *os.File, remoteFilePath string) error {
	c, err := a.sftpClient()
	if err != nil {
		return err
	}

	d, err := createRemote(c, remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file handler at path %s. %+v", remoteFilePath, err)
	}
	defer closeRemoteFile(d)

	if _, err := writeToRemote(d, localSource); err != nil {
		return fmt.Errorf("failed to write to remote file %s. %+v", remoteFilePath, err)
	}

	return nil
}
