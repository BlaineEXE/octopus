// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/BlaineEXE/octopus/internal/remote"
	remotetest "github.com/BlaineEXE/octopus/internal/remote/test"
	"github.com/stretchr/testify/assert"
)

// Octopus needs to:
// 1. get addresses from the host groups file (can fail)
// 2. For each address:
//    1. call remoteConnector.Connect(address) (returns actor) (can fail)
//    2. issue hostname command with actor.RunCommand("hostname")
//    3. call action(actor)
// 3. return the number of hosts that encountered errors
// 4. close actor unless step 2 failed
func TestOctopus_Do(t *testing.T) {

	allConnects := []string{"1.1.1.1", "2.2.2.2", "10.10.10.10"}
	allHostnames := []string{"1.1.1.1-hostname", "2.2.2.2-hostname", "10.10.10.10-hostname"}
	failGetAddrsFromGroupsFile := false
	getAddrsFromGroupsFile = func(hostGroups []string, groupsFile string) ([]string, error) {
		assert.Equal(t, "cars", hostGroups[0])
		assert.Equal(t, "trucks", hostGroups[1])
		assert.Equal(t, "_test-groups-file", groupsFile)
		if failGetAddrsFromGroupsFile {
			return []string{}, fmt.Errorf("test getaddrsfromhostgroups fail")
		}
		return allConnects, nil
	}

	failActions := 0 // fail this many actions
	actionsRun := 0  // num of actions that have been run
	actorsCalled := []remote.Actor{}
	var testAction remote.Action = func(a remote.Actor) (stdout, stderr *bytes.Buffer, err error) {
		actionsRun++
		actorsCalled = append(actorsCalled, a)
		if actionsRun <= failActions {
			return bytes.NewBufferString("stdout fail"), bytes.NewBufferString("stderr fail"),
				fmt.Errorf("action(actor) fail")
		}
		return bytes.NewBufferString("stdout okay"), bytes.NewBufferString("stderr okay"), nil
	}

	type wants struct {
		connects        []string
		connectFails    []string
		hostnamesClosed []string
		numHostErrors   int
		err             bool
	}
	tests := []struct {
		name                       string
		remoteConnector            *remotetest.MockRemoteConnector
		failGetAddrsFromGroupsFile bool
		failHostname               bool
		failActions                int
		wants                      wants
	}{
		{"fail get addrs from groups file", &remotetest.MockRemoteConnector{}, true, false, 0,
			wants{numHostErrors: -1, err: true}},
		{"fail get hostname", &remotetest.MockRemoteConnector{}, false, true, 0,
			wants{connects: allConnects, hostnamesClosed: allHostnames, err: false}},
		{"fail 1 action", &remotetest.MockRemoteConnector{}, false, false, 1,
			wants{connects: allConnects, hostnamesClosed: allHostnames, numHostErrors: 1, err: false}},
		{"fail 3 actions", &remotetest.MockRemoteConnector{}, false, false, 3,
			wants{connects: allConnects, hostnamesClosed: allHostnames, numHostErrors: 3, err: false}},
		{"fail 1 connect", &remotetest.MockRemoteConnector{
			ErrorOnConnectHost: "2.2.2.2"}, false, false, 0,
			wants{
				connects:        allConnects,
				connectFails:    []string{"2.2.2.2"},
				hostnamesClosed: []string{"1.1.1.1-hostname", "10.10.10.10-hostname"},
				numHostErrors:   1,
				err:             false}},
		{"fail 3 connects", &remotetest.MockRemoteConnector{
			ErrorOnConnectHost: "1"}, false, false, 0,
			wants{
				connects:        allConnects,
				connectFails:    []string{"1.1.1.1", "10.10.10.10"},
				hostnamesClosed: []string{"2.2.2.2-hostname"},
				numHostErrors:   2,
				err:             false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failActions = 0
			actionsRun = 0
			actorsCalled = []remote.Actor{}

			o := &Octopus{
				remoteConnector: tt.remoteConnector,
				hostGroups:      []string{"cars", "trucks"},
				groupsFile:      "_test-groups-file",
			}
			tt.remoteConnector.ReturnActor = &remotetest.MockRemoteActor{}
			tt.remoteConnector.ReturnActor.HostnameError = tt.failHostname
			tt.remoteConnector.HostConnects = []string{}
			tt.remoteConnector.HostConnectFails = []string{}
			failActions = tt.failActions
			failGetAddrsFromGroupsFile = tt.failGetAddrsFromGroupsFile

			gotNumHostErrors, err := o.Do(testAction)
			if (err != nil) != tt.wants.err {
				t.Errorf("Octopus.Do() error = %v, want err %v", err, tt.wants.err)
				return
			}
			// num host errors should equal failActions unless -1 (don't care)
			if tt.wants.numHostErrors >= 0 {
				assert.Equal(t, tt.wants.numHostErrors, gotNumHostErrors)
			}

			assert.ElementsMatch(t, tt.wants.connects, tt.remoteConnector.HostConnects, "connects")
			assert.ElementsMatch(t, tt.wants.connectFails, tt.remoteConnector.HostConnectFails, "connect fails")
			// Connector should return all expected actors except those that fail connection
			successfulConns := len(tt.wants.connects) - len(tt.wants.connectFails)
			assert.Equal(t, successfulConns, len(actorsCalled))
			assert.Equal(t, successfulConns, len(tt.remoteConnector.ActorsReturned))
			for _, a := range tt.remoteConnector.ActorsReturned {
				assert.Equal(t, 1, a.CloseCalled)
			}
		})
	}
}
