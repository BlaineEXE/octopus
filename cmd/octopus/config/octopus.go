package config

import (
	"log"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/octopus"
)

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

	hostGroups := viper.GetString("host-groups")
	if hostGroups == "" {
		os.Stderr.WriteString("ERROR: Required value 'host-groups' was not set in the config or in commandline\n")
		os.Stderr.WriteString(OctopusCmd.UsageString())
		os.Exit(-1)
	}

	groupsFile := getAbsFilePath(viper.GetString("groups-file"))
	identityFile := getAbsFilePath(viper.GetString("identity-file"))

	return octopus.New(
		hostGroups,
		groupsFile,
		identityFile,
	)
}

func getAbsFilePath(path string) string {
	p, err := homedir.Expand(path) // can expand '~' but can't expand '$HOME'
	if err != nil {
		log.Fatalf("Error parsing path %s: %+v", path, err)
	}
	p = os.ExpandEnv(p) // can expand '$HOME' but cannot expand '~'
	a, err := filepath.Abs(p)
	if err != nil {
		log.Fatalf("Could not get absolute path for %s: %+v", path, err)
	}
	return a
}
