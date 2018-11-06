package ssh

import (
	"bytes"
	"fmt"

	"github.com/BlaineEXE/octopus/internal/logger"
	"golang.org/x/crypto/ssh"
)

var newSession = func(c *ssh.Client) (*ssh.Session, error) {
	return c.NewSession()
}

var closeSession = func(s *ssh.Session) error {
	return s.Close()
}

var runCommand = func(s *ssh.Session, command string) error {
	return s.Run(command)
}

// RunCommand runs the command on the Actor's remote host.
func (a *Actor) RunCommand(command string) (stdout, stderr *bytes.Buffer, err error) {
	logger.Info.Println("establishing client connection to host:", a.host)
	session, err := newSession(a.sshClient)
	if err != nil {
		err = fmt.Errorf("failed to run command on host %s: %+v", a.host, err)
		return
	}
	defer closeSession(session)

	logger.Info.Println("running user command on host:", a.host)
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	session.Stdout = stdout
	session.Stderr = stderr

	err = runCommand(session, command)
	return
}
