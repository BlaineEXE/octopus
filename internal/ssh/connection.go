package ssh

import (
	"fmt"
	"io/ioutil"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/remote"
	"golang.org/x/crypto/ssh"
)

// A Connector is able to make SSH connections to remote hosts.
type Connector struct {
	clientConfig *ssh.ClientConfig
	port         uint16
}

// NewConnector returns a new SSH connector.
func NewConnector() *Connector {
	return &Connector{
		clientConfig: &ssh.ClientConfig{
			User:            "root",
			Auth:            []ssh.AuthMethod{},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
		port: 22,
	}
}

// AddIdentityFile adds the identity file's key to the remote authentication methods ssh will try
// when connecting to remote hosts.
func (c *Connector) AddIdentityFile(filePath string) error {
	logger.Info.Println("adding identity file:", filePath)

	key, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("unable to parse private key from file %s. %+v", filePath, err)
	}

	c.clientConfig.Auth = append(c.clientConfig.Auth, ssh.PublicKeys(signer))
	return nil
}

// Port sets the port on which remote connections will be made to hosts. The default port is 22.
func (c *Connector) Port(p uint16) error {
	c.port = p
	return nil
}

// User sets the user on hosts to which connections will be made. The default user is 'root'.
func (c *Connector) User(u string) error {
	c.clientConfig.User = u
	return nil
}

var dialHost = ssh.Dial

// Connect connects to the host via ssh with the options that have been previously set and returns
// an actor which can be called to perform tasks on the remote host.
func (c *Connector) Connect(host string) (remote.Actor, error) {
	if len(c.clientConfig.Auth) == 0 {
		return nil, fmt.Errorf(
			"cannot connect to host %s. no ssh authorization methods have been specified", host)
	}
	logger.Info.Println("dialing host:", host)
	client, err := dialHost("tcp", fmt.Sprintf("%s:%d", host, c.port), c.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial host %s. %+v", host, err)
	}
	a := newActor(host, client)
	return a, nil
}
