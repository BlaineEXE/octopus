package octopus

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Result is the result of an action. The result includes the hostname of the target to
// better help the user identify in human-readable format which host the result is from. The result
// also includes information needed to report success and failure conditions.
type Result struct {
	Hostname string
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	Err      error
}

// Print outputs a result in a nice human readable format, printing main output to stdout
// and error output to stderr.
func (r *Result) Print() {
	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n")
	fmt.Printf(" %s\n", r.Hostname)
	fmt.Printf("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\n\n")
	// if buffer is nil, (*bytes.Buffer).String() returns "<nil>"; do not print this
	o := strings.TrimRight(r.Stdout.String(), "\n")
	if r.Stdout != nil && o != "" {
		fmt.Printf("%s\n\n", o)
	}
	o = strings.TrimRight(r.Stderr.String(), "\n")
	if r.Stderr != nil && o != "" {
		fmt.Fprintf(os.Stderr, "Stderr:\n\n%s\n\n", o) // to stderr
	}
	if r.Err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n\n", r.Err) // to stderr
	}
}
