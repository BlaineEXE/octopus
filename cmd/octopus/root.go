package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var octopusCmd = &cobra.Command{
	Use:   "octopus [flags] <command to execute on remote hosts>",
	Short: "Octopus runs a command on multiple remote hosts in parallel",
	Long: `
    Octopus is a simple pdsh-inspired commandline tool for running the same
    command on multiple remote hosts in parallel. Hosts are grouped together
    into "host groups" in a file which inspired by pdsh's "genders" file. The
    host groups file for Octopus is actually a Bash file with groups defined by
    variable definitions so that the same file may be used easily by both
    Octopus and by user-made scripts.

    Under the hood, Octopus uses ssh connections, and some ssh arguments are
    reflected in Octopus's arguments.

    WARNING: Octopus does not do verification of remote hosts
    (StrictHostKeyChecking=no) and does not add entries to the known hosts file.

    See https://github.com/BlaineEXE/octopus for more documentation.`,
	// Support exactly one arg, which is the command
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			Info.SetOutput(os.Stderr)
		}

		o := octopus{
			command:      args[0],
			hostGroups:   hostGroups,
			groupsFile:   getAbsFilePath(groupsFile),
			identityFile: getAbsFilePath(identityFile),
		}

		numErrs, err := o.Run()
		if err != nil {
			log.Fatalf("%v", err)
		}
		os.Exit(numErrs)
	},
}

// Arguments
var (
	command, hostGroups, groupsFile, identityFile string
	verbose                                       bool
)

func init() {
	octopusCmd.PersistentFlags().StringVarP(&hostGroups, "host-groups", "g", "",
		"comma-separated list of host groups; the command will be run on each host in every group (required)")
	if err := octopusCmd.MarkPersistentFlagRequired("host-groups"); err != nil {
		log.Fatalf("%v", err)
	}

	octopusCmd.PersistentFlags().StringVarP(&groupsFile, "groups-file", "f", defaultGroupsFile,
		"file which defines groups of remote hosts available for execution")

	octopusCmd.PersistentFlags().StringVarP(&identityFile, "identity-file", "i", "$HOME/.ssh/id_rsa",
		"file from which the identity (private key) for public key authentication is read")

	octopusCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"print additional information about octopus progress")
}

func getAbsFilePath(path string) string {
	// This can expand '$HOME' to '/home/username' but cannot expand '~' to '/home/username'
	return os.ExpandEnv(path)
}
