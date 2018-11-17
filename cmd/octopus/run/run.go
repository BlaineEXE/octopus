package run

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/tentacle"
)

const (
	aboutText = "Run the given command on remote hosts."
)

// RunCmd is the 'run' command definition which runs a command on remote hosts.
var RunCmd = &cobra.Command{
	Use:   "run [flags] <COMMAND>",
	Short: aboutText,
	Long:  fmt.Sprintf("\n%s", aboutText),
	// Octopus could support more than one arg here and use all args as one command string, but this
	// is to prevent users from being able  to accidentally shoot themselves in the foot with pipes.
	// You must surround your command in quotes to run a command with pipes on remote hosts,
	// otherwise the pipe will indicate the end of the octopus command and pipe octopus's output to
	// whatever comes after.
	Args: cobra.ExactArgs(1), // support exactly one arg, which is the command
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info.Println("Running command: ", args[0])

		o, err := config.TrainOctopus()
		if err != nil {
			return err
		}

		numErrs, err := o.Do(tentacle.CommandRunner(args[0]))
		if err != nil {
			return fmt.Errorf("octopus run command failure: %+v", err)
		}
		os.Exit(numErrs)
		return nil
	},
}
