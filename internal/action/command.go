package action

import (
	"bytes"
	"fmt"
	"math"

	"github.com/BlaineEXE/octopus/internal/logger"
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
func (r *CommandRunner) Do(context *Context) (*Data, error) {
	data := &Data{
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	}

	logger.Info.Println("establishing client connection to host:", context.Host)
	session, err := context.Client.NewSession()
	if err != nil {
		return data, fmt.Errorf("failed to run command on host %s: %+v", context.Host, err)
	}
	defer session.Close()

	logger.Info.Println("running user command on host:", context.Host)
	session.Stdout = data.Stdout
	session.Stderr = data.Stderr

	return data, session.Run(r.Command)
}
