// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"fmt"
	"log"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	"github.com/BlaineEXE/octopus/cmd/octopus/copy"
	"github.com/BlaineEXE/octopus/cmd/octopus/run"
	"github.com/BlaineEXE/octopus/cmd/octopus/version"
)

func init() {
	octopusCmd := config.OctopusCmd

	// Subcommands
	octopusCmd.AddCommand(version.VersionCmd)
	octopusCmd.AddCommand(run.RunCmd)
	octopusCmd.AddCommand(copy.CopyCmd)
}

func main() {
	if err := config.OctopusCmd.Execute(); err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println("")
}
