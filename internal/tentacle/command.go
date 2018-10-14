package tentacle

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
	"golang.org/x/crypto/ssh"
)

// Command is a tentacle which executes a command on a remote host.
type Command struct {
	Command string
}

// Do executes the command tentacle's command on the remote host.
func (c *Command) Do(target *Target, out chan<- Result) {
	runFailedText := "run command failed"
	result := Result{
		// fallback hostname includes the raw host (e.g., IP) for some ability to identify the host
		Hostname: fmt.Sprintf("%s: could not get hostname", target.Host),
		Stdout:   new(bytes.Buffer),
		Err:      fmt.Errorf("%s: unable to get more detail", runFailedText), // fallback error
	}
	defer func() { out <- result }()

	logger.Info.Println("dialing host: ", target.Host)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", target.Host), target.ClientConfig)
	if err != nil {
		result.Err = fmt.Errorf("%+v: %+v", result.Err, err)
		return
	}

	// get the host's hostname for easier identification
	logger.Info.Println("running hostname command on host: ", target.Host)
	hch := make(chan string)
	go func() {
		b := new(bytes.Buffer)
		err := doRunCommand("hostname", client, b)
		if err != nil {
			// command run wasn't a failure if there were problems getting the hostname
			hch <- result.Hostname // just output whatever hostname was set as the fallback on err
		} else {
			hch <- strings.TrimRight(b.String(), "\n")
		}
		close(hch)
	}()

	logger.Info.Println("running user command on host: ", target.Host)
	result.Err = doRunCommand(c.Command, client, result.Stdout)
	if result.Err != nil {
		result.Err = fmt.Errorf("%s: %+v", runFailedText, result.Err)
	}

	result.Hostname = <-hch
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
		return fmt.Errorf("%+v:\n\n%s", err, strings.TrimRight(stderr.String(), "\n"))
	}
	return err
}
