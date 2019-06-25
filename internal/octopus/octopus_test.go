// Package octopus defines an octopus config struct, and a method to run the octopus command defined
// with the configs provided.
package octopus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/remote"
	remotetest "github.com/BlaineEXE/octopus/internal/remote/test"
	"github.com/stretchr/testify/assert"
)

func TestOctopus_ValidHostGroups(t *testing.T) {
	logger.Info.SetOutput(os.Stderr)
	testRootDir, err := ioutil.TempDir("", "octopus-test-valid-host-groups")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(testRootDir)
	groupFile := path.Join(testRootDir, "_test-groups-file")

	createTestGroupFile := func(groupFileText string) {
		f, err := os.Create(groupFile)
		assert.NoError(t, err)
		f.WriteString(groupFileText)
		f.Close()
	}

	createTestGroupFile(`
export one="1.1.1.1"
export two="2.2.2.2"
export three='3.3.3.3'
export first="$one"
intermediate="$two"
rest="$intermediate $three"
export rest`)
	o := &Octopus{
		remoteConnector: &remotetest.MockRemoteConnector{},
		hostGroups:      []string{},
		groupsFile:      groupFile,
	}
	groups, err := o.ValidHostGroups()
	assert.NoError(t, err)
	assert.ElementsMatch(t, groups, []string{"one", "two", "three", "first", "rest"})

	createTestGroupFile(`
export a="1.1.1.1"
export a0="2.2.2.2"
export _a="3.3.3.3"
export a_0="4.4.4.4"
export A="$a"
export A0="$a0"
export _A="$_a"
export A_0="$a_0"
export __SOME_LONG_complicated_VAR_0_with_lots_Of_sTUff_going_on=hi`)
	o = &Octopus{
		remoteConnector: &remotetest.MockRemoteConnector{},
		hostGroups:      []string{},
		groupsFile:      groupFile,
	}
	groups, err = o.ValidHostGroups()
	assert.NoError(t, err)
	assert.ElementsMatch(t, groups, []string{
		"a",
		"a0",
		"_a",
		"a_0",
		"A",
		"A0",
		"_A",
		"A_0",
		"__SOME_LONG_complicated_VAR_0_with_lots_Of_sTUff_going_on",
	})

	createTestGroupFile(`
export a="1.1.1.1"
export 0a='will fail'
export aaa="2.2.2."`)
	groups, err = o.ValidHostGroups()
	assert.NoError(t, err)
	assert.ElementsMatch(t, groups, []string{"a", "aaa"})

	createTestGroupFile(`
one="1.1.1.1"
two="2.2.2.2"`)
	o = &Octopus{
		remoteConnector: &remotetest.MockRemoteConnector{},
		hostGroups:      []string{},
		groupsFile:      groupFile,
	}
	groups, err = o.ValidHostGroups()
	assert.NoError(t, err)
	assert.Empty(t, groups)
}

func BenchmarkOctopus_ValidHostGroups(b *testing.B) {
	testRootDir, err := ioutil.TempDir("", "octopus-bench-valid-host-groups")
	assert.NoError(b, err)
	defer os.RemoveAll(testRootDir)

	groupFile := path.Join(testRootDir, "_test-groups-file")
	f, err := os.Create(groupFile)
	assert.NoError(b, err)
	f.WriteString(`
export one="1.1.1.1"
export two="2.2.2.2"
export three="3.3.3.3"
export first="$one"
rest="$two"
export rest="$rest $three"
`)
	f.Close()

	for i := 0; i < b.N; i++ {
		o := &Octopus{
			remoteConnector: &remotetest.MockRemoteConnector{},
			hostGroups:      []string{},
			groupsFile:      groupFile,
		}
		o.ValidHostGroups()
	}
}

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
