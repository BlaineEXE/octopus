package ssh

import (
	"fmt"
	"io"
	"sync"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPOptions is defines options for SSH's SFTP subsystem.
type SFTPOptions struct {
	BufferSizeKib   uint16
	RequestsPerFile uint16
}

// UserSFTPOptions changes how the SFTP subsystem will be configured for copying files.
var UserSFTPOptions = SFTPOptions{
	BufferSizeKib:   32,
	RequestsPerFile: 64,
}

func newActor(host string, s *ssh.Client) *Actor {
	return &Actor{
		host:            host,
		sshClient:       s,
		sftpOptions:     UserSFTPOptions,
		_sftpCreateOnce: sync.Once{},
		closers:         []io.Closer{s},
	}
}

// An Actor is able to perform actions on a remote via an SSH connection established to the host.
type Actor struct {
	host        string
	sshClient   *ssh.Client
	sftpOptions SFTPOptions

	// SFTP client creation is done lazily if files are to be copied, and only once for each actor
	_sftpClient     *sftp.Client
	_sftpClientErr  error
	_sftpCreateOnce sync.Once

	// Keep a record of all the things we need to close on Actor.Close()
	closers []io.Closer
}

var newSFTPClient = sftp.NewClient

// lazily instantiate an SFTP client so it doesn't get created if files aren't being copied.
func (a *Actor) sftpClient() (*sftp.Client, error) {
	a._sftpCreateOnce.Do(func() {
		logger.Info.Println("establishing SFTP connection to host:", a.host)
		sftp, err := newSFTPClient(a.sshClient,
			sftp.MaxPacketUnchecked(int(a.sftpOptions.BufferSizeKib)*1024), // also convert kib to bytes
			sftp.MaxConcurrentRequestsPerFile(int(a.sftpOptions.RequestsPerFile)))
		a._sftpClient = sftp
		a._sftpClientErr = err
		if err == nil {
			a.closers = append(a.closers, sftp) // make sure SFTP client is closed on Actor.Close
		} else {
			err = fmt.Errorf("failed to start SFTP subsystem. %+v", err)
		}
	})
	return a._sftpClient, a._sftpClientErr
}

// Close closes the SSH connection and (if it exists) SFTP connection.
func (a *Actor) Close() error {
	closeErrs := []error{}
	for _, closer := range a.closers {
		switch closer {
		case a.sshClient:
			logger.Info.Println("closing SSH client for host", a.host)
		case a._sftpClient:
			logger.Info.Println("closing SFTP client for host", a.host)
		default:
			logger.Info.Println("closing unknown closer for host", a.host)
		}
		if err := closer.Close(); err != nil {
			closeErrs = append(closeErrs, err)
		}
	}
	if len(closeErrs) == 0 {
		return nil
	}
	err := fmt.Errorf("error closing Actor for host %s", a.host)
	for e := range closeErrs {
		err = fmt.Errorf("%+v. %+v", err, e) // append each error
	}
	return err
}
