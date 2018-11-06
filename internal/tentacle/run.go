package tentacle

import (
	"bytes"

	"github.com/BlaineEXE/octopus/internal/remote"
)

// CommandRunner returns a new remote action definition which defines how actions are to be run
// on an actor's remote host.
func CommandRunner(command string) remote.Action {
	return func(a remote.Actor) (stdout, stderr *bytes.Buffer, err error) {
		return a.RunCommand(command)
	}
}
