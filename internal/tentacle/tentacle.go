package tentacle

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// Tentacle represents one of Octopus's many parallel-running processes that perform an action
// on a remote host.
// A tentacle can "Do" an action. An action takes a target and returns a result.
type Tentacle interface {
	Do(target *Target, result chan<- Result)
}

// Target is the target of a tentacle action. It defines the remote host the tentacle is to act on,
// and it contains information necessary for the tentacle to understand how to handle that host.
type Target struct {
	Host         string
	ClientConfig *ssh.ClientConfig
}

// Result is the result of a tentacle action. The result includes the hostname of the target to
// better help the user identify in human-readable format which host the result is from. The result
// also includes information needed to report success and failure conditions.
type Result struct {
	Hostname string
	Stdout   *bytes.Buffer
	Err      error
}

// Print outputs a tentacle result in a nice human readable format, printing main output to stdout
// and error output to stderr.
func (r *Result) Print() {
	fmt.Printf("-----\n")
	fmt.Printf("%s\n", r.Hostname)
	fmt.Printf("-----\n\n")
	o := strings.TrimRight(r.Stdout.String(), "\n")
	if o != "" {
		fmt.Printf("%s\n\n", o)
	}
	if r.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n\n", r.Err) // to stderr
	}
}
