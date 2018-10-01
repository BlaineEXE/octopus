// Package logger provides a global info logger for Octopus
package logger

import (
	"io/ioutil"
	"log"
)

var (
	// Info is the logger used for debug printing
	Info *log.Logger
)

func init() {
	Info = log.New(ioutil.Discard, "INFO: ", 0) // Don't output info messages by default
}
