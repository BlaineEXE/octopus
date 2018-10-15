// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"fmt"
	"strings"

	"github.com/BlaineEXE/octopus/internal/action"
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

// RunCommand politely asks the octopus to run a command.
func (o *Octopus) RunCommand(command string) (numHostErrors int, err error) {
	c := &action.RunCommand{
		Command: command,
	}
	return o.exec(c)
}

// CopyFiles prays, "I beg thee Sir Octopus, wilst thou copy these files yonder?"
func (o *Octopus) CopyFiles(localSources []string, remoteDest string) (numHostErrors int, err error) {
	c := &action.CopyFiles{
		LocalSources: localSources,
		RemoteDir:    remoteDest,
	}
	return o.exec(c)
}

func (o *Octopus) exec(action action.Doer) (numHostErrors int, err error) {
	logger.Info.Println("TODO: MORE PRINT INFO HERE")

	g := strings.Split(o.HostGroups, ",")
	logger.Info.Println("host groups: ", g)
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
