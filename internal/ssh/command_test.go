package ssh

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/crypto/ssh"
)

func TestActor_RunCommand(t *testing.T) {
	runtimeNewSession := newSession
	runtimeRunCommand := runCommand
	runtimeCloseSession := closeSession
	defer func() { newSession = runtimeNewSession }()
	defer func() { runCommand = runtimeRunCommand }()
	defer func() { closeSession = runtimeCloseSession }()

	stubClient := &ssh.Client{}
	stubSession := &ssh.Session{}

	stubNewSession := func(c *ssh.Client) (*ssh.Session, error) {
		assert.Equal(t, stubClient, c)
		return stubSession, nil
	}
	failNewSession := func(c *ssh.Client) (*ssh.Session, error) {
		stubNewSession(c)
		return nil, fmt.Errorf("test new session error")
	}

	stubRunCommand := func(s *ssh.Session, command string) error {
		assert.Equal(t, stubSession, s)
		s.Stdout.Write([]byte(command + " stdout okay"))
		s.Stderr.Write([]byte(command + " stderr okay"))
		return nil
	}
	failRunCommand := func(s *ssh.Session, command string) error {
		s.Stdout.Write([]byte(command + " stdout fail"))
		s.Stderr.Write([]byte(command + " stderr fail"))
		return fmt.Errorf("test run command error")
	}

	var sessionClosed bool
	closeSession = func(s *ssh.Session) error {
		sessionClosed = true
		assert.Equal(t, stubSession, s)
		return nil
	}

	tests := []struct {
		name           string
		newSession     func(c *ssh.Client) (*ssh.Session, error)
		runCommand     func(s *ssh.Session, command string) error
		command        string
		wantSessClosed bool
		wantStdout     string
		wantStderr     string
		wantErr        bool
	}{
		{"successful run", stubNewSession, stubRunCommand, "watermelon",
			true, "watermelon stdout okay", "watermelon stderr okay", false},
		{"fail connect", failNewSession, stubRunCommand, "apple",
			false, "<nil>", "<nil>", true},
		{"fail command", stubNewSession, failRunCommand, "banana",
			true, "banana stdout fail", "banana stderr fail", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Actor{
				host:      "test-host",
				sshClient: stubClient,
			}
			sessionClosed = false
			newSession = tt.newSession
			runCommand = tt.runCommand
			gotStdout, gotStderr, err := a.RunCommand(tt.command)
			if (err != nil) != tt.wantErr {
				t.Errorf("Actor.RunCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantStdout, gotStdout.String())
			assert.Equal(t, tt.wantStderr, gotStderr.String())
			assert.Equal(t, tt.wantSessClosed, sessionClosed)
		})
	}
}
