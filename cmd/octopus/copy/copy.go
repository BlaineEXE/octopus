package copy

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/BlaineEXE/octopus/internal/logger"
)

const (
	aboutText = "Copy local files to a dir on remote hosts."
)

// CopyCmd is the 'copy' command definition which copies local files to remote hosts.
var CopyCmd = &cobra.Command{
	// TODO: Add optional '--to' flag
	Use:   "copy [flags] LOCAL_SOURCE_PATHS... REMOTE_DEST_DIR",
	Short: aboutText,
	Long:  fmt.Sprintf("\n%s", aboutText),
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := len(args)
		localSources := args[:n-1]
		remoteDir := args[n-1]
		logger.Info.Println("copying", len(localSources), "local sources", localSources, "to remote dir", remoteDir)

		o := config.TrainOctopus()

		numErrs, err := o.CopyFiles(localSources, remoteDir)
		if err != nil {
			return fmt.Errorf("octopus copy files failure: %+v", err)
		}
		os.Exit(numErrs)
		return nil
	},
}
