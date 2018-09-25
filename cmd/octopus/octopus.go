// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if err := octopusCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Println(groupsFile)

	g := strings.Split(hostGroups, ",")
	hostAddrs, err := getAddrsFromHostsFile(g, groupsFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	config, err := newCommandConfig(identityFile)
	if err != nil {
		log.Fatalf("could not generate command config: %v", err)
	}

	tch := make(chan tentacle, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		go runCommand(hostAddrs[i], command, config, tch)
	}

	numErrors := 0
	for range hostAddrs {
		t := <-tch
		err := t.print()
		if err != nil {
			numErrors++
		}
	}

	os.Exit(numErrors)
}
