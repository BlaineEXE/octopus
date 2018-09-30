// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

var (
	// Info is the logger used for debug printing
	Info *log.Logger
)

func main() {
	Info = log.New(ioutil.Discard, "INFO: ", 0) // Don't output info messages by default

	if err := octopusCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println("")
}
