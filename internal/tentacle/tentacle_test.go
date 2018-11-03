package tentacle

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/BlaineEXE/octopus/internal/action"
	"golang.org/x/crypto/ssh"
)

func TestTentacle_Go(t *testing.T) {
	type fields struct {
		Host   string
		Action action.Doer
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
			Action:       &mockAction{&expectedContext{t, host, testClient}},
			ClientConfig: testClientConfig} // shared input client config
	}

	// convenience func for reducing verbosity of setting a mock below
	// failHostnameGetter is as simple as &failHostnamegetter{}
	mockHostnameGetter := func(host string) *mockHostnameGetter {
		return &mockHostnameGetter{&expectedContext{t, host, testClient}}
	}

	tests := []struct {
		name           string
		tentacle       *Tentacle
		sshDialer      func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
		hostnameGetter action.Doer
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
type expectedContext struct {
	t      *testing.T
	host   string
	client *ssh.Client
}

func (ct *expectedContext) test(c *action.Context) {
	if c.Host != ct.host {
		ct.t.Errorf("context.Host = %s, want %s", c.Host, ct.host)
	}
	if c.Client != ct.client { // compare pointers is fine
		ct.t.Errorf("context.Client = %v, want %v", c.Client, ct.client)
	}
}

type failHostnameGetter struct{}

func (*failHostnameGetter) Do(_ *action.Context) (*action.Data, error) {
	return nil, fmt.Errorf("test failure on commandrunner.Do get hostname")
}

// mock hostnameGetter tests that it gets the right context
type mockHostnameGetter struct {
	*expectedContext
}

func (h *mockHostnameGetter) Do(context *action.Context) (*action.Data, error) {
	h.test(context)
	return &action.Data{
		Stdout: bytes.NewBufferString("test hostname"),
		Stderr: bytes.NewBufferString("should not appear"),
	}, nil
}

type failAction struct{}

func (*failAction) Do(context *action.Context) (*action.Data, error) {
	return &action.Data{
		Stdout: bytes.NewBufferString("stdout-fail"),
		Stderr: bytes.NewBufferString("stderr-fail"),
	}, fmt.Errorf("test failure on tentacle.Action.Do")
}

// mock action tests that it gets the right context
type mockAction struct {
	*expectedContext
}

func (m *mockAction) Do(context *action.Context) (*action.Data, error) {
	m.test(context)
	return &action.Data{
		Stdout: bytes.NewBufferString("stdout-good"),
		Stderr: bytes.NewBufferString("stderr-good"),
	}, nil
}

type expectedResult struct {
	hostnameContains string // hostname should contain this string, "" is don't care
	stdout           string // Data.Stdout should stringify to this, "" is don't care
	stderr           string // Data.Stderr should stringify to this, "" is don't care
	wantErr          bool   // error desired in result
}

func (e *expectedResult) test(t *testing.T, r *Result) {
	if !strings.Contains(r.Hostname, e.hostnameContains) {
		t.Errorf("Result.Hostname = %s, want %s", r.Hostname, e.hostnameContains)
	}
	if r.Data != nil {
		if e.stdout != "" && e.stdout != r.Data.Stdout.String() {
			t.Errorf("Result.Data.Stdout = %s, want %s", r.Data.Stdout.String(), e.stdout)
		}
		if e.stderr != "" && e.stderr != r.Data.Stderr.String() {
			t.Errorf("Result.Data.Stderr = %s, want %s", r.Data.Stderr.String(), e.stdout)
		}
	}
	if e.wantErr != (r.Err != nil) {
		t.Errorf("Result.Err? %t, wanted? %t", (r.Err != nil), e.wantErr)
	}
}
