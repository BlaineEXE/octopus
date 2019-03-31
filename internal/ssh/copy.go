package ssh

import (
	"fmt"
	"os"
	"time"

	"github.com/pkg/sftp"
)

var statRemote = func(c *sftp.Client, dirPath string) (os.FileInfo, error) {
	return c.Stat(dirPath)
}

var mkdirAllRemote = func(c *sftp.Client, dirPath string, mode os.FileMode) error {
	if err := c.MkdirAll(dirPath); err != nil {
		return err
	}
	return c.Chmod(dirPath, mode)
}

// CreateRemoteDir creates the dir as well as any nonexistent parents on the Actor's remote host if
// any of the dirs do not exist. Return nil if the paths already exist.
func (a *Actor) CreateRemoteDir(dirPath string, perms os.FileMode) error {
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
	} else if err := mkdirAllRemote(c, dirPath, perms); err != nil {
		return fmt.Errorf(errMsg, err)
	}
	return nil
}

var createRemote = func(c *sftp.Client, filePath string, info os.FileInfo) (*sftp.File, error) {
	r, err := c.Create(filePath)
	if err != nil {
		return nil, err
	}
	r.Chmod(info.Mode().Perm())
	return r, nil
}

var writeToRemote = func(dest *sftp.File, source *os.File) (int64, error) {
	return dest.ReadFrom(source)
}

var closeRemoteFile = func(f *sftp.File) error {
	return f.Close()
}

// CopyFileToRemote copies the file to the Actor's remote host at the remote file path.
func (a *Actor) CopyFileToRemote(localSource *os.File, remoteFilePath string, info os.FileInfo) error {
	c, err := a.sftpClient()
	if err != nil {
		return err
	}

	d, err := createRemote(c, remoteFilePath, info)
	if err != nil {
		return fmt.Errorf("failed to create remote file handler at path %s. %+v", remoteFilePath, err)
	}
	defer closeRemoteFile(d)

	if _, err := writeToRemote(d, localSource); err != nil {
		return fmt.Errorf("failed to write to remote file %s. %+v", remoteFilePath, err)
	}

	if err := c.Chtimes(remoteFilePath, time.Now(), info.ModTime()); err != nil {
		return fmt.Errorf("failed to set the remote file %s's last modified time. %+v", remoteFilePath, err)
	}

	return nil
}
