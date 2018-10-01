// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
)

// Octopus is a metaphorical octopus which can run commands on remote hosts in parallel with its
// many arms.
type Octopus struct {
	HostGroups   string
	GroupsFile   string
	IdentityFile string
}

// New finds an octopus and trains it about how its environment is configured and what host groups
// it should operate on.
func New(HostGroups, GroupsFile, IdentityFile string) *Octopus {
	return &Octopus{
		HostGroups:   HostGroups,
		GroupsFile:   GroupsFile,
		IdentityFile: IdentityFile,
	}
}

// Run politely asks the octopus to run a command.
func (o *Octopus) Run(command string) (numHostErrors int, err error) {
	logger.Info.Println("user command:\n", command)

	g := strings.Split(o.HostGroups, ",")
	logger.Info.Println("host groups: ", g)
	hostAddrs, err := getAddrsFromGroupsFile(g, o.GroupsFile)
	if err != nil {
		return -1, err
	}

	config, err := newCommandConfig(o.IdentityFile)
	if err != nil {
		return -1, fmt.Errorf("could not generate command config: %v", err)
	}

	tch := make(chan tentacle, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		go runCommand(hostAddrs[i], command, config, tch)
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

// Marshal marshalls the octopus to a string.
func Marshal(octopus *Octopus) (string, error) {
	j, err := json.Marshal(octopus)
	if err != nil {
		return "", fmt.Errorf("failed to marshal the octopus %v to a string: %v", octopus, err)
	}
	return string(j), nil
}

// Unmarshal unmarshalls the string to an octopus.
func Unmarshal(marshalledOctopus string) (*Octopus, error) {
	var o Octopus
	m := []byte(marshalledOctopus)
	if err := json.Unmarshal(m, &o); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the string %v to an octopus: %v", marshalledOctopus, err)
	}
	return &o, nil
}
