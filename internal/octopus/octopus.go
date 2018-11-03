// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"fmt"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/tentacle"
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

// Do sends out tentacles to all hosts in the host group(s) in individual goroutines and collects
// the results of all the tentacles at the end. Returns the number of hosts that report errors if
// the tentacles are able to be sent out.
func (o *Octopus) Do(action tentacle.RemoteDoer) (numHostErrors int, err error) {
	g := strings.Split(o.HostGroups, ",")
	logger.Info.Println("host groups:", g)
	hostAddrs, err := getAddrsFromGroupsFile(g, o.GroupsFile)
	if err != nil {
		return -1, err
	}

	config, err := newClientConfig(o.IdentityFile)
	if err != nil {
		return -1, fmt.Errorf("could not generate ssh client config: %+v", err)
	}

	rch := make(chan tentacle.Result, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		tntcl := &tentacle.Tentacle{
			Host:         hostAddrs[i],
			Action:       action,
			ClientConfig: config,
		}
		go tntcl.Go(rch)
	}

	numHostErrors = 0
	for range hostAddrs {
		r := <-rch
		r.Print()
		if r.Err != nil {
			numHostErrors++
		}
	}
	return numHostErrors, nil
}
