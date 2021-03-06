package root

import (
	"github.com/BlaineEXE/octopus/cmd/octopus/completion"
	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/BlaineEXE/octopus/cmd/octopus/copy"
	"github.com/BlaineEXE/octopus/cmd/octopus/hostgroups"
	"github.com/BlaineEXE/octopus/cmd/octopus/run"
	"github.com/BlaineEXE/octopus/cmd/octopus/version"
)

func init() {
	octopusCmd := config.OctopusCmd

	// Subcommands
	octopusCmd.AddCommand(completion.BashCompletionCmd)
	octopusCmd.AddCommand(version.VersionCmd)
	octopusCmd.AddCommand(hostgroups.HostGroupsCommand)
	octopusCmd.AddCommand(run.RunCmd)
	octopusCmd.AddCommand(copy.CopyCmd)
}
