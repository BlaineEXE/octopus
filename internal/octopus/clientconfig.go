package octopus

import (
	"fmt"
	"io/ioutil"

	"github.com/BlaineEXE/octopus/internal/logger"
	"golang.org/x/crypto/ssh"
)

// Allow this to be overridden for testing.
var parsePrivateKey = ssh.ParsePrivateKey

var newClientConfig = func(identityFile string) (*ssh.ClientConfig, error) {
	logger.Info.Println("identity file: ", identityFile)

	key, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %+v", err)
	}

	signer, err := parsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %+v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return config, nil
}
