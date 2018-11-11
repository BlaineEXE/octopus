package ssh

import (
	"fmt"
	"io"
	"path"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/BlaineEXE/octopus/internal/remote"
	"github.com/BlaineEXE/octopus/internal/util/testutil"
	"golang.org/x/crypto/ssh"
)

const (
	parsableKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAyi+522d48Wk/w4240N2BT8WYHhJiPeDYbDs3BkvpDYbkqKiG
XKIs988UXFlev/dEN1iJj4MAwGWkEKCotkd329ECTa8I5ijiVsYEMJS4kIub2TLO
+JFlkkADohffoe89TobQcscV+WW8Mb6kqfn5l2DbwjwO+1hBarcpe5W6rbqttLsa
HGafQ1gpZIYNjyDDhg+G/BcNSWEhqAaVOO7kGmiI/oMSrnGaGHo50DIyo9k73jhv
e4WS8fh52/65uP2bf/BqqVVaTeLOR/xw8VcsIu7qpna97UlDHV0/cpfY4oXBSaA2
sPVcHfmLMQgKPjSX33pL8VIWbp9dikxTuUrcCQIDAQABAoIBACoIvBISNAG0iO4l
86twsaadBOAToFsO+M+fi/QCKiSGy9kImE5/5OtsIOaGKf2s8YC0Jn0wliJpvy85
x3mF6DRKArmVzzrSeyPhLRPQh4J9k9wXBRKGX+CE8HxtjC/FZjCCNKn4G+hqrzKE
WQWBk9NV7ro19ENq+Mav567q1RGy47F1E92Czfn1i30xz0h9kSd1fTyy0W1511YI
0gIjpkUyKlJHzXH3PREsSoQ0sBI9LdyBhyiX11d1xsiIaVPTei/HJYNgJu4U98FJ
DQkpBwms04WefQHt6CgJZXmwAa8IM3H14GG3vMLDdlCgFOAKDzYQ9nlTiv+e3wGB
blQUiiUCgYEA/Rb/7NipySkxztMQ6QFbyD/Oe7OxZodELirykI6K9g62dIvKj5LZ
DPoJeynNNyg4ADS5h7P4zMx+52pQMTFG9G8qkfLijaWmEodmGO5hGBCVxc5F0H5p
OdLmRirpOD7lj0Q3xAjpYpc5N2S4459CJwM1pC117WBeBdODNOx3vAsCgYEAzILi
0MYtUR48bCQLLsT0nkmHskaccA7Xn499DEHRN70v4ZjEIP78eHF7JBBJDoR+Vwsn
UUI0VeBis4aSS8y12CzWW0lPOeRYWzWvFTZ17wtODgHr2O0tJQsUPyxpBq5zTK3+
3blGJ7zlnzv29vhAmS3iYPZ1XBVg8+ZZ5IHNgLsCgYEA6Q4N207kgh3SBM9tifK7
TtoazOR6npw+13iq5xyrr8t6jjXP8IfcIUv6ARVMKNd7Qg1LL0A2Anjo/zZx4+qp
mRrpC36qyp7YH8XY6WpRtHRJRt4cgdJ2GU4wyDppima4w0WhSH6gUy5H/M9eRhT4
OK6G7ckDB/SugBT2hHygAWMCgYBHtpnchbG8YTLk5Nq7AruYicY4oIQY00uPGxzJ
YIcB2ahhnlUgEOntPjXlFoTXv2QiF7oox2Ncvbs+orDIPbeCX26nQhSzAzxsd222
rYs7UKaFSO0v+zM6ayElaehGPIQX3mehzmcoZhfK95cJUVItpKZeQ+4xZRnDTQI2
m8G5IwKBgEOLL4kJR3M6xY+e7CnXJe0q8VHgjTHaos1SwkxMbqyjGImYlI9VbS+8
iXj7N8nBMSiMXGwcJsZU3YsoBDkNGBfn3aOgUeOp6wvXJJYVQmNDju8zKoL3Vh1U
cyRWx1gHb95Ie5yyrBnuRgOWjrTvGjrI7rdJ49LpdiXSucd5m72F
-----END RSA PRIVATE KEY-----
`
	unparsableKey = `
blah blah blah
`
)

func TestConnector_AddIdentityFile(t *testing.T) {
	tmpRoot, cleanup := testutil.TempDir("")
	defer cleanup()
	parsableKeyfile := path.Join(tmpRoot, "parsableKey")
	unparsableKeyfile := path.Join(tmpRoot, "unparsableKey")
	writeonlyFile := path.Join(tmpRoot, "writeonlyKey")
	testutil.WriteFile(parsableKeyfile, parsableKey, 0644)
	testutil.WriteFile(unparsableKeyfile, unparsableKey, 0644)
	testutil.WriteFile(writeonlyFile, parsableKey, 0222)

	tests := []struct {
		name        string
		connector   *Connector
		filePath    string // arg
		wantNumAuth int    // number of clientConfig.Auth methods there should be
		wantErr     bool
	}{
		{"add first id file", NewConnector(), parsableKeyfile, 1, false},
		{"unreadable keyfile", NewConnector(), writeonlyFile, 0, true},
		{"unparsable keyfile", NewConnector(), unparsableKeyfile, 0, true},
		{"add second id file",
			func() *Connector { c := NewConnector(); c.AddIdentityFile(parsableKeyfile); return c }(),
			parsableKeyfile, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.connector.AddIdentityFile(tt.filePath); (err != nil) != tt.wantErr {
				t.Errorf("Connector.AddIdentityFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.wantNumAuth, len(tt.connector.clientConfig.Auth))
		})
	}
}

func TestConnector_Connect(t *testing.T) {
	tmpRoot, cleanup := testutil.TempDir("")
	defer cleanup()
	parsableKeyfile := path.Join(tmpRoot, "parsableKey")
	testutil.WriteFile(parsableKeyfile, parsableKey, 0644)

	// Test connector ready to go with an identity file added
	stubConnector := NewConnector()
	stubConnector.AddIdentityFile(parsableKeyfile)

	stubClient := &ssh.Client{} // stub dialer always returns this client

	var lastAddrDialed string // for checking that the right address was dialed
	failDial := true
	dialHost = func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		lastAddrDialed = addr
		if failDial {
			return nil, fmt.Errorf("test dial error")
		}
		return stubClient, nil
	}

	// sort of tests newActor as well
	expectedActor := func(host string) *Actor {
		return &Actor{
			host:            host,
			sshClient:       stubClient,
			_sftpCreateOnce: sync.Once{},             // must init the sync.Once for SFTP
			_sftpClient:     nil,                     // do not create an sftp client on SSH connect
			_sftpClientErr:  nil,                     // do not report SFTP client err on SSH connect
			closers:         []io.Closer{stubClient}, // actor should know to close the stub client
		}
	}

	type wants struct {
		addrDialed string
		actor      remote.Actor
		err        bool
	}
	tests := []struct {
		name      string
		connector *Connector
		host      string // arg
		failDial  bool
		wants     wants
	}{
		{"successful connection", stubConnector, "1.1.1.1", false, wants{
			addrDialed: "1.1.1.1:22", actor: expectedActor("1.1.1.1"), err: false}},
		{"fail to dial", stubConnector, "2.2.2.2", true, wants{
			addrDialed: "2.2.2.2:22", actor: nil, err: true}},
		{"connector w/o id file", NewConnector(), "not dialed", false, wants{
			addrDialed: "not dialed", actor: nil, err: true}},
		{"connect with a different port",
			func() (c *Connector) { c = new(Connector); *c = *stubConnector; c.Port(2222); return }(),
			"3.3.3.3", false, wants{
				addrDialed: "3.3.3.3:2222", actor: expectedActor("3.3.3.3"), err: false}},
	}
	for _, tt := range tests {
		lastAddrDialed = "not dialed"
		t.Run(tt.name, func(t *testing.T) {
			failDial = tt.failDial
			a, err := tt.connector.Connect(tt.host)
			if (err != nil) != tt.wants.err {
				t.Errorf("Connector.Connect() error = %v, want error %v", err, tt.wants.err)
			}
			if lastAddrDialed != tt.wants.addrDialed {
				t.Errorf("Connector.Connect() dialed addr %s, want %s", lastAddrDialed, tt.wants.addrDialed)
			}
			if !reflect.DeepEqual(a, tt.wants.actor) {
				t.Errorf("Connector.Connect() Actor = %+v, want %+v", a, tt.wants.actor)
			}
		})
	}
}
