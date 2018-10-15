package action

import (
	"bytes"

	"golang.org/x/crypto/ssh"
)

// Doer is the interface that wraps the Do method.
type Doer interface {
	Do(context *Context) (*Data, error)
}

// Context is a structure containing information that is contextual to its running environment.
type Context struct {
	Host   string
	Client *ssh.Client
}

// Data is the structure returned by a doer.
type Data struct {
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Err    error
}
