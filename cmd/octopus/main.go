// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"fmt"
	"log"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	_ "github.com/BlaineEXE/octopus/cmd/octopus/root"
)

func main() {
	if err := config.OctopusCmd.Execute(); err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println("")
}
