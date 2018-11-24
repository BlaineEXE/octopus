package config

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/version"
)

const (
	defaultGroupsFile = "_node-list"
)

// OctopusCmd is the top-level 'octopus' command.
var OctopusCmd = &cobra.Command{
	Use:   "octopus [flags] [--host-groups|-h <HOST-GROUPS>] <COMMAND>",
	Short: "Octopus runs a command on multiple remote hosts in parallel",
	Long: `
-----------
  OCTOPUS
-----------

  Octopus is a simple pdsh-inspired commandline tool for running the same
  command on multiple remote hosts in parallel. Hosts are grouped together
  into "host groups" in a file which inspired by a "genders" file. The
  host groups file for Octopus is actually a Bash file with groups defined by
  variable definitions. This is so that the same file may be used easily by
	both Octopus and by user-made scripts and has the secondary benefit of
	supporting defining hosts by IP address as well as hostname.

  Under the hood, Octopus uses ssh connections, and some ssh arguments are
  reflected in Octopus's arguments. These arguments are marked in the help
  text with "(ssh)".

  WARNINGS:
    Octopus does not do verification of remote hosts (Equivalent to
    setting ssh option StrictHostKeyChecking=no) and does not add entries
    to the known hosts file.

  Config file:
    Octopus supports setting custom default values for flags in a config file.
  Any of Octopus's top-level, full-name flags can be set in the config.
  Octopus will search in order for the first 'config.yaml' file it finds in:
    (1) ./.octopus
    (2) $HOME/.octopus
    (3) /etc/octopus
  e.g., Simply writing "host-groups: all" into the config file will use the
  'all' host group for Octopus commands unless the user specifies a different
  set of host groups using '--host-groups|-h' on the commandline.
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("verbose") {
			logger.Info.SetOutput(os.Stderr)
			logger.Info.Println("Running octopus in verbose mode")
			logger.Info.Println("Octopus version:", version.Version)
		}
	},
}

// octopusCmd can't reference itself to print usage when there is an input error
var usageString string

func init() {
	// Load the config file at cobra initialization
	cobra.OnInitialize(loadConfig)

	// Persistent top-level flags
	OctopusCmd.PersistentFlags().StringP("groups-file", "f", defaultGroupsFile,
		"file which defines groups of remote hosts available for execution")
	OctopusCmd.PersistentFlags().StringSliceP("host-groups", "g", []string{},
		"comma-separated list of host groups; the command will be run on each host in every group")
	OctopusCmd.PersistentFlags().StringP("identity-file", "i", "$HOME/.ssh/id_rsa",
		"(ssh) file from which the identity (private key) for public key authentication is read")
	OctopusCmd.PersistentFlags().Uint16P("port", "p", 22,
		"(ssh) port on which to connect to hosts")
	OctopusCmd.PersistentFlags().StringP("user", "u", "root",
		"user as which to connect to hosts (corresponds to ssh \"-l\" option)")
	OctopusCmd.PersistentFlags().BoolP("verbose", "v", false,
		"print additional information about octopus progress")

	// Persistent flags are also valid config file options
	viper.BindPFlags(OctopusCmd.PersistentFlags())
}
