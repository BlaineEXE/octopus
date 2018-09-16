// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	identityFile := flag.String("identity-file", "~/.ssh/id_rsa",
		"identity file used to authenticate to remote hosts")
	command := flag.String("command", "", "(required) command to execute on remote hosts")
	flag.Parse()

	if strings.Trim(*command, " \t") == "" {
		fmt.Printf("ERROR! '-command' must be specified\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := newCommandConfig(*identityFile)
	if err != nil {
		log.Fatalf("could not generate command config: %v", err)
	}

	hosts := []string{"10.86.1.87", "10.86.1.103"}
	tentacles := make(chan tentacle, len(hosts))

	for i := 0; i < len(hosts); i++ {
		go runCommand(hosts[i], *command, config, tentacles)
	}

	numErrors := 0

	for range hosts {
		t := <-tentacles
		err := t.print()
		if err != nil {
			numErrors++
		}
	}

	os.Exit(numErrors)
}
