// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"fmt"
	"log"
)

func main() {
	if err := octopusCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println("")
}
