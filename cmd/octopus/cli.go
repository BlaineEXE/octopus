package main

import (
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var octopusCmd = &cobra.Command{
	Use:   "octopus [flags] [--host-groups|-h <HOST-GROUPS>] <COMMAND>",
	Short: "Octopus runs a command on multiple remote hosts in parallel",
	Long: `
-----------
  OCTOPUS
-----------

  Octopus is a simple pdsh-inspired commandline tool for running the same
  command on multiple remote hosts in parallel. Hosts are grouped together
  into "host groups" in a file which inspired by pdsh's "genders" file. The
  host groups file for Octopus is actually a Bash file with groups defined by
  variable definitions. This is so that the same file may be used easily by
  both Octopus and by user-made scripts.

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
	// Support exactly one arg, which is the command
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("verbose") {
			Info.SetOutput(os.Stderr)
			Info.Println("Config file used: ", viper.ConfigFileUsed())
		}

		hostGroups := viper.GetString("host-groups")
		if hostGroups == "" {
			os.Stderr.WriteString("ERROR: Required value 'host-groups' was not set in the config or in commandline\n")
			// octopusCmd can't reference itself to print usage when there is an input error,
			// so we have to get/set the usage string here circuitously.
			os.Stderr.WriteString(usageString)
			os.Exit(1)
		}

		o := octopus{
			command:      args[0],
			hostGroups:   viper.GetString("host-groups"),
			groupsFile:   getAbsFilePath(viper.GetString("groups-file")),
			identityFile: getAbsFilePath(viper.GetString("identity-file")),
		}

		numErrs, err := o.Run()
		if err != nil {
			log.Fatalf("%v", err)
		}
		os.Exit(numErrs)
	},
}

// octopusCmd can't reference itself to print usage when there is an input error
var usageString string

func init() {
	cobra.OnInitialize(initConfig)

	octopusCmd.PersistentFlags().StringP("host-groups", "g", "",
		"(required) comma-separated list of host groups; the command will be run on each host in every group")

	octopusCmd.PersistentFlags().StringP("groups-file", "f", defaultGroupsFile,
		"file which defines groups of remote hosts available for execution")

	octopusCmd.PersistentFlags().StringP("identity-file", "i", "$HOME/.ssh/id_rsa",
		"(ssh) file from which the identity (private key) for public key authentication is read")

	octopusCmd.PersistentFlags().BoolP("verbose", "v", false,
		"print additional information about octopus progress")

	viper.BindPFlags(octopusCmd.PersistentFlags())

	// octopusCmd can't reference itself to print usage when there is an input error
	usageString = octopusCmd.UsageString()
}

// Read from the config file
func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./.octopus")
	viper.AddConfigPath("$HOME/.octopus")
	viper.AddConfigPath("/etc/octopus/")
	err := viper.ReadInConfig()
	if err != nil && !isConfigFileNotFoundError(err) {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func getAbsFilePath(path string) string {
	p, err := homedir.Expand(path) // can expand '~' but can't expand '$HOME'
	if err != nil {
		log.Fatalf("Error parsing path %s: %v", path, err)
	}
	p = os.ExpandEnv(p) // can expand '$HOME' but cannot expand '~'
	a, err := filepath.Abs(p)
	if err != nil {
		log.Fatalf("Could not get absolute path for %s: %v", path, err)
	}
	return a
}

func isConfigFileNotFoundError(err error) bool {
	return reflect.TypeOf(err) == reflect.TypeOf(viper.ConfigFileNotFoundError{})
}
