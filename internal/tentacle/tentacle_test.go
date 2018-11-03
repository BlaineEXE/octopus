package tentacle

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestTentacle_Go(t *testing.T) {
	type fields struct {
		Host   string
		Action RemoteDoer
	}

	testClientConfig := &ssh.ClientConfig{} // all tests have same client config as input
	testClient := &ssh.Client{}             // mock ssh dialer returns this client

	failSSHDialer := func(_, _ string, c *ssh.ClientConfig) (*ssh.Client, error) {
		fmt.Println("faildial")
		if c != testClientConfig { // dial needs to be called with the tentacle's client config
			t.Errorf("dial called with client config %+v, want %+v", c, testClientConfig)
		}
		return nil, fmt.Errorf("test failure on ssh.Dial")
	}
	mockSSHDialer := func(_, _ string, _ *ssh.ClientConfig) (*ssh.Client, error) {
		return testClient, nil
	}

	failTentacle := func(host string) *Tentacle {
		return &Tentacle{
			Host:         host,
			Action:       &failAction{},
			ClientConfig: testClientConfig} // shared input client config
	}
	mockTentacle := func(host string) *Tentacle {
		return &Tentacle{
			Host:         host,
			Action:       &mockAction{&expectedArgs{t, host, testClient}},
			ClientConfig: testClientConfig} // shared input client config
	}

	// convenience func for reducing verbosity of setting a mock below
	// failHostnameGetter is as simple as &failHostnamegetter{}
	mockHostnameGetter := func(host string) *mockHostnameGetter {
		return &mockHostnameGetter{&expectedArgs{t, host, testClient}}
	}

	tests := []struct {
		name           string
		tentacle       *Tentacle
		sshDialer      func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
		hostnameGetter RemoteDoer
		expectedResult *expectedResult
	}{
		{"fail to dial", failTentacle("1.1.1.1"), failSSHDialer, &failHostnameGetter{},
			&expectedResult{"", "", "", true}},
		// For hostname failures, the host name returned should contain host the addr for manual identification.
		{"hostname and action failure", failTentacle("2.2.2.2"), mockSSHDialer, &failHostnameGetter{},
			&expectedResult{"2.2.2.2", "stdout-fail", "stderr-fail", true}},
		{"hostname failure", mockTentacle("3.3.3.3"), mockSSHDialer, &failHostnameGetter{},
			&expectedResult{"3.3.3.3", "stdout-good", "stderr-good", false}},
		{"action failure", failTentacle("4.4.4.4"), mockSSHDialer, mockHostnameGetter("4.4.4.4"),
			&expectedResult{"test hostname", "stdout-fail", "stderr-fail", true}},
		{"full success", mockTentacle("5.5.5.5"), mockSSHDialer, mockHostnameGetter("5.5.5.5"),
			&expectedResult{"test hostname", "stdout-good", "stderr-good", false}},
	}
	for _, tt := range tests {
		out := make(chan Result)
		defer close(out)
		t.Run(tt.name, func(t *testing.T) {
			tntcl := tt.tentacle
			dialHost = tt.sshDialer
			hostnameGetter = tt.hostnameGetter
			go tntcl.Go(out)
			r := <-out
			tt.expectedResult.test(t, &r)
		})
	}
}

// test that the context sent to the action matches what should be sent to it
type expectedArgs struct {
	t      *testing.T
	host   string
	client *ssh.Client
}

func (e *expectedArgs) test(host string, client *ssh.Client) {
	if host != e.host {
		e.t.Errorf("context.Host = %s, want %s", host, e.host)
	}
	if client != e.client { // compare pointers is fine
		e.t.Errorf("context.Client = %v, want %v", client, e.client)
	}
}

type failHostnameGetter struct{}

func (*failHostnameGetter) Do(_ string, _ *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	err = fmt.Errorf("test failure on commandrunner.Do get hostname")
	return
}

// mock hostnameGetter tests that it gets the right context
type mockHostnameGetter struct {
	*expectedArgs
}

func (h *mockHostnameGetter) Do(host string, client *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	h.test(host, client)
	stdout = bytes.NewBufferString("test hostname")
	stderr = bytes.NewBufferString("should not appear")
	err = nil
	return
}

type failAction struct{}

func (*failAction) Do(host string, client *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	stdout = bytes.NewBufferString("stdout-fail")
	stderr = bytes.NewBufferString("stderr-fail")
	err = fmt.Errorf("test failure on tentacle.Action.Do")
	return
}

// mock action tests that it gets the right context
type mockAction struct {
	*expectedArgs
}

func (m *mockAction) Do(host string, client *ssh.Client) (stdout, stderr *bytes.Buffer, err error) {
	m.test(host, client)
	stdout = bytes.NewBufferString("stdout-good")
	stderr = bytes.NewBufferString("stderr-good")
	err = nil
	return
}

type expectedResult struct {
	hostnameContains string // hostname should contain this string, "" is don't care
	stdout           string // Stdout should stringify to this, "" is don't care
	stderr           string // Stderr should stringify to this, "" is don't care
	wantErr          bool   // error desired in result
}

func (e *expectedResult) test(t *testing.T, r *Result) {
	if !strings.Contains(r.Hostname, e.hostnameContains) {
		t.Errorf("Result.Hostname = %s, want %s", r.Hostname, e.hostnameContains)
	}
	if e.stdout != "" && e.stdout != r.Stdout.String() {
		t.Errorf("Result.Stdout = %s, want %s", r.Stdout.String(), e.stdout)
	}
	if e.stderr != "" && e.stderr != r.Stderr.String() {
		t.Errorf("Result.Stderr = %s, want %s", r.Stderr.String(), e.stdout)
	}
	if e.wantErr != (r.Err != nil) {
		t.Errorf("Result.Err? %t, wanted? %t", (r.Err != nil), e.wantErr)
	}
}
