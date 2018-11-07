// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"fmt"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/remote"
)

// Octopus is a metaphorical octopus which can run commands on remote hosts in parallel with its
// many arms.
type Octopus struct {
	remoteConnector remote.Connector
	hostGroups      []string
	groupsFile      string
}

// New finds an octopus and trains it about how its environment is configured and what host groups
// it should operate on.
func New(c remote.Connector, hostGroups []string, groupsFile string) *Octopus {
	return &Octopus{
		remoteConnector: c,
		hostGroups:      hostGroups,
		groupsFile:      groupsFile,
	}
}

// Do sends out tentacles to all hosts in the host group(s) in individual goroutines and collects
// the results of all the tentacles at the end. Returns the number of hosts that report errors if
// the tentacles are able to be sent out.
func (o *Octopus) Do(action remote.Action) (numHostErrors int, err error) {
	logger.Info.Println("host groups:", o.hostGroups)
	hostAddrs, err := getAddrsFromGroupsFile(o.hostGroups, o.groupsFile)
	if err != nil {
		return -1, err
	}

	rch := make(chan Result, len(hostAddrs))
	for i := 0; i < len(hostAddrs); i++ {
		go func(host string) {
			result := Result{
				// fallback hostname includes the raw host (e.g., IP) for some ability to identify the host
				Hostname: fmt.Sprintf("%s: could not get hostname", host),
				// fallback error - should never be returned, but *just* in case, make sure it isn't nil
				Err: fmt.Errorf("failed to send tentacle: unable to get more detail"),
			}
			defer func() { rch <- result }()
			actor, err := o.remoteConnector.Connect(host)
			if err != nil {
				result.Err = err
				return
			}
			defer actor.Close()

			// get the host's hostname (in parallel) for easier human identification
			logger.Info.Println("running hostname command on host:", host)
			hch := make(chan string)
			go func() {
				defer close(hch)
				o, _, err := actor.RunCommand("hostname")
				if err != nil {
					hch <- result.Hostname // use fallback hostname on error
					return
				}
				hch <- strings.TrimRight(o.String(), "\n")
			}()

			// Do whatever action the user wants
			result.Stdout, result.Stderr, result.Err = action(actor)

			result.Hostname = <-hch
		}(hostAddrs[i])
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
