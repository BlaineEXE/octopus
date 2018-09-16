// Package octopus is a commandline tool for running the same command on multiple remote hosts in
// parallel.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	noHostNameText = "! could not get hostname !"
)

// Each of octopus's tentacles is a remote connection to a host executing the command
type tentacle struct {
	host     string
	hostname string
	stdout   *bytes.Buffer
	err      error
}

func main() {
	identityFile := flag.String("identity-file", "~/.ssh/id_rsa",
		"identity file used to authenticate to remote hosts")
	command := flag.String("command", "", "(required) command to execute on remote hosts")
	flag.Parse()

	if strings.Trim(*command, " \t") == "" {
		fmt.Printf("ERROR! '-command' must be specified\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	key, err := ioutil.ReadFile(*identityFile)
	if err != nil {
		log.Fatalf("unable to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hosts := []string{"10.86.1.87", "10.86.1.103"}
	tentacles := make(chan tentacle, len(hosts))

	for i := 0; i < len(hosts); i++ {
		go runCommand(hosts[i], *command, config, tentacles)
	}

	numErrors := 0

	for range hosts {
		t := <-tentacles
		err := t.print()
		if err != nil {
			numErrors++
		}
	}

	os.Exit(numErrors)
}

func runCommand(host, command string, config *ssh.ClientConfig, out chan<- tentacle) {
	t := tentacle{
		host:     host,
		hostname: "",
		err:      fmt.Errorf("run command failed"),
	}
	defer func() { out <- t }()

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", t.host), config)
	if err != nil {
		t.err = fmt.Errorf("%v: %v", t.err, err)
		return
		// log.Fatal("Failed to dial: ", err)
	}

	hn := make(chan tentacle)
	go getHostname(client, hn)

	session, err := client.NewSession()
	if err != nil {
		t.err = fmt.Errorf("%v: %v", t.err, err)
		// log.Fatal("Failed to create session: ", err)
	}
	defer session.Close()

	t.stdout = new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	session.Stdout = t.stdout
	session.Stderr = stderr
	err = session.Run(command)

	if err != nil {
		t.err = fmt.Errorf("%v: %v\n\n%s\n\n%s", t.err, err, strings.TrimRight(stderr.String(), "\n"), "")
	} else {
		t.err = nil
	}

	tn := <-hn
	t.hostname = tn.hostname
	return
}

func getHostname(client *ssh.Client, out chan<- tentacle) {
	defer close(out)

	t := tentacle{
		hostname: noHostNameText,
	}
	defer func() { out <- t }()

	session, err := client.NewSession()
	if err != nil {
		t.err = fmt.Errorf("%v: %v", t.err, err)
		return
	}
	defer session.Close()

	t.stdout = new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	session.Stdout = t.stdout
	session.Stderr = stderr
	err = session.Run("hostname")
	if err == nil {
		t.hostname = strings.TrimRight(t.stdout.String(), "\n")
	}

	return
}

func (t *tentacle) print() error {
	fmt.Println("-----")
	fmt.Println(t.hostname)
	if t.hostname == noHostNameText {
		fmt.Println(t.host)
	}
	fmt.Printf("-----\n\n")
	o := strings.TrimRight(t.stdout.String(), "\n")
	if o != "" {
		fmt.Printf("%s\n\n", o)
	}
	if t.err != nil {
		fmt.Printf("Error: %v", t.err)
	}
	return t.err
}
