package test

import (
	"fmt"
	"strings"
	"sync"

	"github.com/BlaineEXE/octopus/internal/remote"
)

// MockRemoteConnector is a reusable mock remote.Connector to be used for testing.
type MockRemoteConnector struct {
	// Config
	ErrorOnIdentityFile string
	ErrorOnConnectHost  string
	ReturnActor         *MockRemoteActor

	// Results
	IdentityFileAdds     []string
	IdentityFileAddFails []string
	HostConnects         []string
	HostConnectFails     []string
	ActorsReturned       []*MockRemoteActor
}

// Connector may be called in parallel
var connectorMutex = sync.Mutex{}

// AddIdentityFile is a mock method that appends each filePath to IdentityFileAdds.
// If filePath contains ErrorOnIdentityFile, an error will be returned, and filePath appended to IdentityFileAddFails.
func (c *MockRemoteConnector) AddIdentityFile(filePath string) error {
	connectorMutex.Lock()
	defer connectorMutex.Unlock()
	app(&c.IdentityFileAdds, filePath)
	if c.ErrorOnIdentityFile != "" && strings.Contains(filePath, c.ErrorOnIdentityFile) {
		app(&c.IdentityFileAddFails, filePath)
		return fmt.Errorf(filePath + " fail")
	}
	return nil
}

// Connect is a mock method that appends each host to HostConnects.
// It returns a copy of ReturnActor with Hostname="host-hostname"
// If host contains ErrorOnHostConnect, an error will be returned, and host appended to HostConnectFails.
func (c *MockRemoteConnector) Connect(host string) (remote.Actor, error) {
	connectorMutex.Lock()
	defer connectorMutex.Unlock()
	app(&c.HostConnects, host)
	if c.ErrorOnConnectHost != "" && strings.Contains(host, c.ErrorOnConnectHost) {
		app(&c.HostConnectFails, host)
		return nil, fmt.Errorf(host + " fail")
	}
	r := &MockRemoteActor{}
	*r = *c.ReturnActor
	r.Hostname = host + "-hostname"
	//fmt.Println("r hostname:", r.Hostname)
	c.ActorsReturned = append(c.ActorsReturned, r)
	return r, nil
}
