package ssh

import (
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/pkg/sftp"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

type stubCloser struct {
	returnErr   string
	closeCalled int
}

func (s *stubCloser) Close() error {
	s.closeCalled++
	if s.returnErr != "" {
		return fmt.Errorf(s.returnErr)
	}
	return nil
}
func (s *stubCloser) reset() {
	s.closeCalled = 0
}

func TestActor_Close(t *testing.T) {
	closer1 := &stubCloser{}
	closer2 := &stubCloser{}
	failCloser1 := &stubCloser{returnErr: "failCloser1"}
	failCloser2 := &stubCloser{returnErr: "failCloser2"}
	allClosers := []*stubCloser{closer1, closer2, failCloser1, failCloser2}

	tests := []struct {
		name         string
		actorClosers []*stubCloser
		wantErr      bool
	}{
		{"1 good closer, 0 fail closers", []*stubCloser{closer1}, false},
		{"2 good closers, 0 fail closers", []*stubCloser{closer1, closer2}, false},
		{"0 good closers, 1 fail closer", []*stubCloser{failCloser1}, true},
		{"0 good closers, 2 fail closers", []*stubCloser{failCloser2, failCloser1}, true},
		{"1 good closer, 1 fail closer", []*stubCloser{closer1, failCloser2}, true},
		{"1 fail closer, 1 good closer", []*stubCloser{failCloser1, closer2}, true},
		{"2 good closers, 2 fail closers", []*stubCloser{closer2, closer1, failCloser1, failCloser2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Actor{
				closers: []io.Closer{},
			}
			for _, c := range tt.actorClosers {
				a.closers = append(a.closers, c)
			}
			if err := a.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Actor.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			// If the closer is in our actor, make sure close has been called exactly once,
			// otherwise make sure it definitely has not been called. And reset the close counter.
			for _, c := range allClosers {
				if isIn(tt.actorClosers, c) {
					assert.Equal(t, 1, c.closeCalled)
				} else {
					assert.Zero(t, c.closeCalled)
				}
				c.reset()
			}
		})
	}
}

func isIn(cs []*stubCloser, key io.Closer) bool {
	for _, c := range cs {
		if c == key {
			return true
		}
	}
	return false
}

func TestActor_sftpClient(t *testing.T) {
	stubSSHClient := &ssh.Client{}

	runtimeNewSFTPClient := newSFTPClient
	defer func() { newSFTPClient = runtimeNewSFTPClient }()

	stubSFTPClient := &sftp.Client{}
	stubNewSFTPClient := func(conn *ssh.Client, opts ...sftp.ClientOption) (*sftp.Client, error) {
		if conn != stubSSHClient {
			t.Errorf("Actor.sftpClient() called newSFTPClient with wrong ssh.Client: %+v, expected %+v", conn, stubSSHClient)
		}
		return stubSFTPClient, nil
	}
	failNewSFTPClient := func(conn *ssh.Client, opts ...sftp.ClientOption) (*sftp.Client, error) {
		stubNewSFTPClient(conn, opts...)
		return nil, fmt.Errorf("test sftp connect error")
	}

	tests := []struct {
		name          string
		newSFTPClient func(conn *ssh.Client, opts ...sftp.ClientOption) (*sftp.Client, error)
		want          *sftp.Client
		wantErr       bool
	}{
		{"successful connect", stubNewSFTPClient, stubSFTPClient, false},
		{"failed connect", failNewSFTPClient, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newActor(tt.name, stubSSHClient)
			newSFTPClient = tt.newSFTPClient
			s1, e1 := a.sftpClient()
			if (e1 != nil) != tt.wantErr {
				t.Errorf("Actor.sftpClient() error = %v, wantErr %v", e1, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(s1, tt.want) {
				t.Errorf("Actor.sftpClient() = %v, want %v", s1, tt.want)
			}
			// output and errors should be the same after the second call to sftpClient()
			s2, e2 := a.sftpClient()
			assert.Equal(t, s1, s2)
			assert.Equal(t, e1, e2)
		})
	}
}
