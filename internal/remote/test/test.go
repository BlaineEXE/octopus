package test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// MockRemoteActor is a reusable mock remote.Actor to be used for testing.
type MockRemoteActor struct {
	// Config
	Hostname         string // hostname to return
	HostnameError    bool   // issue error on hostname command?
	CommandError     bool   // issue error on command?
	CreateDirErrorOn string // issue error when dir contains this string ("" is no error)
	CopyFileErrorOn  string // issue error when file contains this string ("" is no error)

	// Results
	Commands       []string // all commands actor has attempted to run
	DirCreates     []string // all dirs actor has attempted to create (incl. failed ones)
	DirCreateFails []string // dirs actor has failed to create
	FileCopies     []string // all files actor has attempted to copy (incl. failed ones)
	FileCopyFails  []string // files actor has failed to copy
	CloseCalled    int      // Close has been called this many times
}

// Clear takes a pointer to a result list as a param, returns the contents, and empties the list.
// e.g., Clear(&m.Commands)
func Clear(list *[]string) (contents []string) {
	contents = *list
	*list = []string{}
	return
}

// shortcut for bytes.NewBufferString
var bs = bytes.NewBufferString

// append to string slice in-place
func app(ss *[]string, s string) {
	*ss = append(*ss, s)
}

// RunCommand is a mock function that appends each command to Commands.
// It will return Hostname if the command is "hostname", or an error if HostnameError is true.
// It will always return data on stdout and stderr in the form below where command is the command
// intput, stdout/stderr is the buffer on which the data is returned, and ok unless CommandError is
// true, in which case err:
//   <command>: <stdout|stderr> <ok|err>
func (m *MockRemoteActor) RunCommand(command string) (stdout, stderr *bytes.Buffer, err error) {
	app(&m.Commands, command)

	if command == "hostname" {
		if m.HostnameError {
			return bs(""), bs("hostnameerror"), fmt.Errorf("test hostname error")
		}
		return bs(m.Hostname), bs(""), nil
	}
	if m.CommandError {
		return bs(command + ": stdout err"), bs(command + ": stderr err"),
			fmt.Errorf("test command error running command %s", command)
	}
	return bs(command + ": stdout ok"), bs(command + ": stderr ok"), nil
}

// ExpectedCommandOutput returns string versions of stdout and stderr expected for the command
// and the command's expected error state.
func ExpectedCommandOutput(command string, err bool) (stdout, stderr string) {
	okTxt := "ok"
	if err {
		okTxt = "err"
	}
	stdout = command + ": stdout " + okTxt
	stderr = command + ": stderr " + okTxt
	return
}

// CreateRemoteDir is a mock function that appends each remote dir path to DirCreates.
// It will return an error after it has been called CreateDirErrorAfter number of times.
func (m *MockRemoteActor) CreateRemoteDir(dirPath string) error {
	app(&m.DirCreates, dirPath)

	if m.CreateDirErrorOn != "" && strings.Contains(dirPath, m.CreateDirErrorOn) {
		app(&m.DirCreateFails, dirPath)
		return fmt.Errorf("test error creating remote dir %s", dirPath)
	}
	return nil
}

// CopyFileToRemote is a mock function that appends each remote file path to FileCopies.
// It will return an error after it has been called CopyFileErrorAfter number of times.
func (m *MockRemoteActor) CopyFileToRemote(localSource *os.File, remoteFilePath string) error {
	app(&m.FileCopies, remoteFilePath)

	if m.CopyFileErrorOn != "" && strings.Contains(remoteFilePath, m.CopyFileErrorOn) {
		app(&m.FileCopyFails, remoteFilePath)
		return fmt.Errorf("test error copying file to remote at %s", remoteFilePath)
	}
	return nil
}

// Close is a mock function that increments CloseCalled each time it is called.
// It does not return an error.
func (m *MockRemoteActor) Close() error {
	m.CloseCalled++
	return nil
}
