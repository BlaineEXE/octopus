// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
//
// See config for a sample host groups file
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	command := flag.String("command", "", "(required) command to execute on remote hosts")
	hostGroups := flag.String("host-groups", "",
		"(required) comma-separated list of host groups on which to execute the command")
	hostsFile := flag.String("hosts-file", defaultHostsFile, fmt.Sprintf(
		"file which defines which groups of remote hosts are available for execution (default: %s)", defaultHostsFile))
	identityFile := flag.String("identity-file", "~/.ssh/id_rsa",
		"identity file used to authenticate to remote hosts")
	flag.Parse()

	if strings.Trim(*command, " \t") == "" {
		fmt.Printf("ERROR! '-command' must be specified\n\n")
		flag.PrintDefaults()
		os.Exit(-1)
	}
	if strings.Trim(*hostGroups, " \t") == "" {
		fmt.Printf("ERROR! '-hosts' must be specified \n\n")
		flag.PrintDefaults()
		os.Exit(-1)
	}

	h := strings.Split(*hostGroups, ",")
	hostAddrs, err := getAddrsFromHostsFile(h, *hostsFile)
	if err != nil {
		log.Fatalf("%v", err)
	}

	config, err := newCommandConfig(*identityFile)
	if err != nil {
		log.Fatalf("could not generate command config: %v", err)
	}

	tch := make(chan tentacle, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		go runCommand(hostAddrs[i], *command, config, tch)
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
