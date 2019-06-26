package tentacle

import (
	"testing"

	remotetest "github.com/BlaineEXE/octopus/internal/remote/test"
	"github.com/stretchr/testify/assert"
)

func TestCommandRunner(t *testing.T) {
	a := remotetest.MockRemoteActor{}

	type args struct {
		command string
	}
	tests := []struct {
		name   string
		args   args
		cmdErr bool // should the command report error?
	}{
		{"no error", args{"goodcommand"}, false},
		{"error", args{"badcommand"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.args.command
			a.CommandError = tt.cmdErr

			action := CommandRunner(cmd)
			o, e, err := action(&a)
			assert.True(t, (err != nil) == tt.cmdErr) //err received when expected

			run := remotetest.Clear(&a.Commands)
			assert.Equal(t, 1, len(run))
			assert.Equal(t, cmd, run[0])
			assert.Zero(t, len(a.DirCreates))
			assert.Zero(t, len(a.FileCopies))

			eso, ese := remotetest.ExpectedCommandOutput(cmd, tt.cmdErr)
			assert.Contains(t, o.String(), eso)
			assert.Contains(t, e.String(), ese)

		})
	}
}
