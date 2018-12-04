package main

import (
	"os"

	"github.com/BlaineEXE/octopus/cmd/octopus/config"
	_ "github.com/BlaineEXE/octopus/cmd/octopus/root"
)

func main() {
	config.OctopusCmd.GenBashCompletion(os.Stdout)
}
