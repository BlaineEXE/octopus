// Package version defines the 'octopus version' command behavior. It's one responsibility is to
// report Octopus's version to the user.
package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BlaineEXE/octopus/internal/version"
)

const (
	aboutText = "Print Octopus's version information"
)

// VersionCmd is the 'version' command definition which prints version information.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: aboutText,
	Long:  fmt.Sprintf("\n%s", aboutText),
	Args:  cobra.ExactArgs(0), // support no args to version
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(" octopus version ", version.Version)
		return nil
	},
}
