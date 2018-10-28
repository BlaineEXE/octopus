package tentacle

import (
	"fmt"
	"os"
	"strings"

	"github.com/BlaineEXE/octopus/internal/action"
	"github.com/BlaineEXE/octopus/internal/logger"
	"golang.org/x/crypto/ssh"
)

// Tentacle represents one of Octopus's many parallel-running processes that perform an action
// on a remote host.
// A tentacle can "Do" an action. An action takes a target and returns a result.
type Tentacle struct {
	Host         string
	Action       action.Doer
	ClientConfig *ssh.ClientConfig
}

// Result is the result of a tentacle action. The result includes the hostname of the target to
// better help the user identify in human-readable format which host the result is from. The result
// also includes information needed to report success and failure conditions.
type Result struct {
	Hostname string
	Data     *action.Data
	Err      error
}

// Go sends out a tentacle to start a new remote connection and do the action on the remote host.
func (t *Tentacle) Go(out chan<- Result) {
	result := Result{
		// fallback hostname includes the raw host (e.g., IP) for some ability to identify the host
		Hostname: fmt.Sprintf("%s: could not get hostname", t.Host),
		Data:     &action.Data{},
		// fallback error - should never be returned, but *just* in case, make sure it isn't nil
		Err: fmt.Errorf("failed to send tentacle: unable to get more detail"),
	}
	defer func() { out <- result }()

	logger.Info.Println("dialing host: ", t.Host)
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", t.Host), t.ClientConfig)
	if err != nil {
		result.Err = fmt.Errorf("failed to start tentacle to host %s (failed to dial): %+v", t.Host, err)
		return
	}

	context := &action.Context{
		Host:   t.Host,
		Client: client,
	}

	// get the host's hostname (in parallel) for easier human identification
	logger.Info.Println("running hostname command on host:", t.Host)
	hch := make(chan string)
	hostnameGetter := action.RunCommand{Command: "hostname"}
	go func() {
		defer close(hch)
		hData, err := hostnameGetter.Do(context)
		if err != nil {
			hch <- result.Hostname // use fallback hostname on error
			return
		}
		hch <- strings.TrimRight(hData.Stdout.String(), "\n")
	}()

	// Do whatever action the user wants
	result.Data, result.Err = t.Action.Do(context)

	result.Hostname = <-hch
	return
}

// Print outputs a tentacle result in a nice human readable format, printing main output to stdout
// and error output to stderr.
func (r *Result) Print() {
	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n")
	fmt.Printf(" %s\n", r.Hostname)
	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	// if buffer is nil, (*bytes.Buffer).String() returns "<nil>"; do not print this
	o := strings.TrimRight(r.Data.Stdout.String(), "\n")
	if r.Data.Stdout != nil && o != "" {
		fmt.Printf("%s\n\n", o)
	}
	o = strings.TrimRight(r.Data.Stderr.String(), "\n")
	if r.Data.Stderr != nil && o != "" {
		fmt.Fprintf(os.Stderr, "Stderr:\n\n%s\n\n", o) // to stderr
	}
	if r.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n\n", r.Err) // to stderr
	}
}
