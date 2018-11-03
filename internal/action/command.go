package action

import (
	"bytes"
	"fmt"
	"math"

	"github.com/BlaineEXE/octopus/internal/logger"
	"golang.org/x/crypto/ssh"
)

// CommandRunner is a tentacle action which executes a command on a remote host.
type CommandRunner struct {
	Command string
}

// JobLimit is the command tentacle's limit on the number of command jobs that can be started.
// As many command tentacles can be started as desired, as there are no shared resource limitations.
func (r *CommandRunner) JobLimit() uint32 {
	// run command doesn't have any resource limits of its own, so return what is basically infinity
	return math.MaxUint32
}

// Do executes the command tentacle's command on the remote host.
func (r *CommandRunner) Do(host string, client *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	logger.Info.Println("establishing client connection to host:", host)
	session, err := client.NewSession()
	if err != nil {
		err = fmt.Errorf("failed to run command on host %s: %+v", host, err)
		return
	}
	defer session.Close()

	logger.Info.Println("running user command on host:", host)
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)
	session.Stdout = stdout
	session.Stderr = stderr

	err = session.Run(r.Command)
	return
}
