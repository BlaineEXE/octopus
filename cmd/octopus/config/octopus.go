package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/octopus"
	"github.com/BlaineEXE/octopus/internal/remote"
	"github.com/BlaineEXE/octopus/internal/ssh"
	"github.com/BlaineEXE/octopus/internal/util"
)

var remoteConnector remote.Connector = ssh.NewConnector()

// TrainOctopus returns an octopus trained (configured) for the user's environment.
// This sanitizes values set in the config file and on the commandline to ensure the octopus
// is well-trained.
// This should be called by subcommands which wish to use an octopus, and it should *not* be called
// by subcommands which will not use an octopus since octopus-specific inputs are validated to be
// present in this function.
// If a required input for an octopus action is not present, this function will print CLI usage
// information and exit the program with a failure code.
func TrainOctopus() (*octopus.Octopus, error) {
	logger.Info.Println("Config file used: ", viper.ConfigFileUsed())
	logger.Info.Println("Parsing global flags")

	hostGroups := viper.GetStringSlice("host-groups")
	if len(hostGroups) == 0 {
		// host-groups is not required for 'version' command; only commands that require an octopus
		// to be created. Do a manual check here so that Cobra doesn't check 'version' for
		// host-groups. Do not return an error here, but instead print to stderr and exit nonzero
		// to control what the output looks like more exactly.
		os.Stderr.WriteString("ERROR: Required value 'host-groups' was not set in the config or in commandline\n")
		os.Stderr.WriteString(OctopusCmd.UsageString())
		os.Exit(-1)
	}
	logger.Info.Println("Host groups:", hostGroups)

	groupsFile := getAbsFilePath(viper.GetString("groups-file"))
	identityFile := getAbsFilePath(viper.GetString("identity-file"))

	if err := remoteConnector.AddIdentityFile(identityFile); err != nil {
		return nil, fmt.Errorf("could not add identity file: %+v", err)
	}
	if err := remoteConnector.Port(uint16(viper.GetInt("port"))); err != nil {
		return nil, fmt.Errorf("could not change port: %+v", err) // ssh always return nil here
	}
	if err := remoteConnector.User(viper.GetString("user")); err != nil {
		return nil, fmt.Errorf("could not change user: %+v", err) // ssh always return nil here
	}

	return octopus.New(
		remoteConnector,
		hostGroups,
		groupsFile,
	), nil
}

func getAbsFilePath(path string) string {
	a, err := util.AbsPath(path)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return a
}
