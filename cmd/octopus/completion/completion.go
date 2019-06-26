package completion

import (
	"os"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/spf13/cobra"
)

// BashCompletionCmd is the 'completion' command definition which produces Bash completion
// scripts for Octopus.
var BashCompletionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Output Bash completion code for Octopus",
	Long: `
  Output Bash completion code for Octopus. Octopus completion is limited to
  Bash only and follows the same patterns as completion for Kubernetes' kubectl.
  kubectl's autocompletion install docs should provide an excellent overview
  applicable to Octopus as well:
  https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion
`,
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.OctopusCmd.GenBashCompletion(os.Stdout)
	},
}
