package config

import (
	"log"
	"os"
	"strings"

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
func TrainOctopus() *octopus.Octopus {
	logger.Info.Println("Config file used: ", viper.ConfigFileUsed())
	logger.Info.Println("Parsing global flags")

	rawHostGroups := viper.GetString("host-groups")
	if rawHostGroups == "" {
		os.Stderr.WriteString("ERROR: Required value 'host-groups' was not set in the config or in commandline\n")
		os.Stderr.WriteString(OctopusCmd.UsageString())
		os.Exit(-1)
	}
	hostGroups := strings.Split(rawHostGroups, ",")

	groupsFile := getAbsFilePath(viper.GetString("groups-file"))
	identityFile := getAbsFilePath(viper.GetString("identity-file"))

	remoteConnector.AddIdentityFile(identityFile)

	return octopus.New(
		remoteConnector,
		hostGroups,
		groupsFile,
	)
}

func getAbsFilePath(path string) string {
	a, err := util.AbsPath(path)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return a
}
