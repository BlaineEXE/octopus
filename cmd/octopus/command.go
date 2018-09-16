package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"golang.org/x/crypto/ssh"
)

func newCommandConfig(identityFile string) (*ssh.ClientConfig, error) {
	key, err := ioutil.ReadFile(identityFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v", err)
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

func runCommand(host, command string, config *ssh.ClientConfig, out chan<- tentacle) {
	runFailedText := "run command failed"
	t := tentacle{
		// fallback hostname includes the raw host (e.g., IP) for some ability to identify the host
		hostname: fmt.Sprintf("%s: could not get hostname", host),
		stdout:   new(bytes.Buffer),
		err:      fmt.Errorf("%s: unable to get more detail", runFailedText), // fallback error
	}
	defer func() { out <- t }()

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		t.err = fmt.Errorf("%v: %v", t.err, err)
		return
	}

	// Get the host's hostname for easier identification
	hch := make(chan string)
	go func() {
		b := new(bytes.Buffer)
		err := doRunCommand("hostname", client, b)
		if err != nil {
			// command run wasn't a failure if there were problems getting the hostname
			hch <- t.hostname // just output whatever hostname was set as the fallback on err
		} else {
			hch <- strings.TrimRight(b.String(), "\n")
		}
		close(hch)
	}()

	t.err = doRunCommand(command, client, t.stdout)
	if t.err != nil {
		t.err = fmt.Errorf("%s: %v", runFailedText, t.err)
	}

	t.hostname = <-hch
	return
}

func doRunCommand(command string, client *ssh.Client, stdoutBuffer *bytes.Buffer) error {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stderr := new(bytes.Buffer)
	session.Stdout = stdoutBuffer
	session.Stderr = stderr

	err = session.Run(command)
	if err != nil {
		return fmt.Errorf("%v:\n\n%s", err, strings.TrimRight(stderr.String(), "\n"))
	}
	return err
}
