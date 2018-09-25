package main

import (
	"fmt"
	"strings"
)

type octopus struct {
	command      string
	hostGroups   string
	groupsFile   string
	identityFile string
}

func (o *octopus) Run() (numHostErrors int, err error) {
	Info.Println("user command:\n", o.command)

	g := strings.Split(o.hostGroups, ",")
	Info.Println("host groups: ", g)
	hostAddrs, err := getAddrsFromGroupsFile(g, o.groupsFile)
	if err != nil {
		return -1, err
	}

	config, err := newCommandConfig(o.identityFile)
	if err != nil {
		return -1, fmt.Errorf("could not generate command config: %v", err)
	}

	tch := make(chan tentacle, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		go runCommand(hostAddrs[i], o.command, config, tch)
	}

	numHostErrors = 0
	for range hostAddrs {
		t := <-tch
		err := t.print()
		if err != nil {
			numHostErrors++
		}
	}
	return numHostErrors, nil
}
