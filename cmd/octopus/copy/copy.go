package copy

import (
	"fmt"
	"os"

	"github.com/BlaineEXE/octopus/internal/ssh"

	"github.com/BlaineEXE/octopus/internal/tentacle"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/BlaineEXE/octopus/internal/logger"
)

// CopyCmd is the 'copy' command definition which copies local files to remote hosts.
var CopyCmd = &cobra.Command{
	// TODO: Add optional '--to' flag
	Use:   "copy [flags] LOCAL_SOURCE_PATHS... REMOTE_DEST_DIR",
	Short: "Copy local files to a dir on remote hosts.",
	Long: `
  Copy local files and/or directories to a given directory on remote hosts. If
  the destination directory does not exist on remote hosts, it (and any
  nonexistent parents) will be created with permissions 0644. Files specified
  individually will be copied with the same permissions as exist locally to the
  destination directory directly. Directories specified will be copied only if
  the 'recursive|r' argument is given, and both permissions and the file tree
  layout within the dir will be copied to the destination dir.

  Copy uses SSH's SFTP subsystem under the hood, and some sftp arguments are
  reflected in Octopus's copy arguments. These arguments are marked in the help
  text with "(sftp)".
   - The default buffer size (--buffer-size|-B) is the guaranteed-to-work
     maximum for all SFTP connections. OpenSSH's maximum is 256 kib but less in
     practice because there is also overhead for TCP packet headers.
   - Octopus's (--requests-per-file|-R) differs somewhat from sftp's -R option
     in that it is a 'per-file' argument in Octopus.
`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		n := len(args)
		localSources := args[:n-1]
		remoteDir := args[n-1]
		logger.Info.Println("copying", len(localSources), "local sources", localSources, "to remote dir", remoteDir)

		o, err := config.TrainOctopus()
		if err != nil {
			return err
		}

		ssh.UserSFTPOptions.BufferSizeKib = uint16(viper.GetInt("buffer-size"))
		ssh.UserSFTPOptions.RequestsPerFile = uint16(viper.GetInt("requests-per-file"))
		logger.Info.Println("SFTP buffer size (kib):", ssh.UserSFTPOptions.BufferSizeKib)
		logger.Info.Println("SFTP requests per file:", ssh.UserSFTPOptions.RequestsPerFile)

		opts := tentacle.NewCopyFileOptions(viper.GetBool("recursive"))
		numErrs, err := o.Do(tentacle.FileCopier(localSources, remoteDir, opts))
		if err != nil {
			return fmt.Errorf("octopus copy files failure: %+v", err)
		}
		os.Exit(numErrs)
		return nil
	},
}

func init() {
	CopyCmd.Flags().BoolP("recursive", "r", false, "recurse into subdirectories and copy all files")
	CopyCmd.Flags().Uint16P("buffer-size", "B", 32,
		"(sftp) in kibibits (kib), maximum buffer (chunk) size for copying files")
	CopyCmd.Flags().Uint16P("requests-per-file", "R", 64, "(sftp) max number of concurrent requests per file")

	viper.BindPFlags(CopyCmd.Flags())
}
