package remote

import (
	"bytes"
	"os"
)

// A Connector can configure how remote connections are to be made and make remote connections to
// hosts with the settings that were configured.
// All remote host connections will share the same configuration.
type Connector interface {
	// AddIdentityFile should add the identity file's key to the remote authentication methods the
	// connector will try when connecting to remote hosts.
	AddIdentityFile(filePath string) error

	// Port should set the port on which remote connections will be made to hosts.
	Port(p uint16) error

	// User should set the user on hosts to which connections will be made.
	User(u string) error

	// Connect should connect to the host with the options that have been previously set and return
	// an actor which can be called to perform tasks on the remote host. If an error is reported,
	// the actor should not need to have its Close method called.
	Connect(host string) (Actor, error)
}

// An Actor can perform a task on a remote host.
// Actors will only be created by calling the Connector.Connect method.
// Multiple Actor tasks will be run simultaneously on each remote host connection.
type Actor interface {
	// RunCommand should run the command on the remote host specified in the Connector.Connect method.
	RunCommand(command string) (stdout, stderr *bytes.Buffer, err error)

	// CreateRemotedir should create a directory along with any nonexistent parents on the remote
	// host specified in the Connector.Connect method. Should return nil if the paths already exist.
	CreateRemoteDir(dirPath string) error

	// CopyFileToRemote should copy the file to the remote host specified in the Connector.Connect
	// method at the remote path, and the remote path includes the remote file name.
	CopyFileToRemote(localSource *os.File, remoteFilePath string) error

	// Close should close all necessary connections the Actor has made.
	Close() error
}

// An Action function is a function that tells an actor how to do a task.
type Action func(a Actor) (stdout, stderr *bytes.Buffer, err error)
