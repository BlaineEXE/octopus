package octopus

import (
	"bytes"
	"fmt"
	"strings"
)

// Each of octopus's tentacles is a remote connection to a host executing the command
type tentacle struct {
	hostname string
	stdout   *bytes.Buffer
	err      error
}

func (t *tentacle) print() error {
	fmt.Println("-----")
	fmt.Println(t.hostname)
	fmt.Printf("-----\n\n")
	o := strings.TrimRight(t.stdout.String(), "\n")
	if o != "" {
		fmt.Printf("%s\n\n", o)
	}
	if t.err != nil {
		fmt.Printf("Error: %v\n\n", t.err)
	}
	return t.err
}
