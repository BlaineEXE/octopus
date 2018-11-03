package octopus

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"golang.org/x/crypto/ssh"
)

var runtimeNewClientConfig func(identityFile string) (*ssh.ClientConfig, error)

func init() {
	// On init, store the default version of the function which will execute at runtime
	// so it isn't lost by tests replacing this with mocks.
	runtimeNewClientConfig = newClientConfig
}

const parsableKey = `
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

const unparsableKey = `
blah blah blah
`

type expectedClientConfig struct {
	t    *testing.T
	user string
}

func (e *expectedClientConfig) test(c *ssh.ClientConfig) {
	if c.User != e.user {
		e.t.Errorf("ClientConfig.User = %s, want %s", c.User, e.user)
	}
}

func Test_newClientConfig(t *testing.T) {
	// Make temp dir for testing
	tmpRoot, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("failed to create temp dir for testing. %+v", err)
	}
	defer os.RemoveAll(tmpRoot)
	parsableKeyfile := path.Join(tmpRoot, "parsableKeyfile")
	unparsableKeyfile := path.Join(tmpRoot, "unparsableKeyfile")
	writeonlyFile := path.Join(tmpRoot, "writeonlyKeyfile")
	// createFile lives in groupsfile_test.go
	createFile(t, parsableKeyfile, parsableKey, false)
	createFile(t, unparsableKeyfile, unparsableKey, false)
	createFile(t, writeonlyFile, parsableKey, true)

	runtimeParsePrivateKey := parsePrivateKey
	defer func() { parsePrivateKey = runtimeParsePrivateKey }()
	parsePrivateKeyAndTest := func(pemBytes []byte) (ssh.Signer, error) {
		if string(pemBytes) != parsableKey {
			t.Errorf("expected key: %s\n    was not passed to parsePrivateKey(): %s", parsableKey, string(pemBytes))
		}
		return runtimeParsePrivateKey(pemBytes)
	}

	type args struct {
		identityFile string
	}
	tests := []struct {
		name            string
		args            args
		parsePrivateKey func(pemBytes []byte) (ssh.Signer, error)
		want            *expectedClientConfig
		wantErr         bool
	}{
		{"unparsable keyfile", args{unparsableKeyfile}, runtimeParsePrivateKey, nil, true},
		{"writeonly keyfile", args{writeonlyFile}, runtimeParsePrivateKey, nil, true},
		{"good keyfile", args{parsableKeyfile}, parsePrivateKeyAndTest,
			&expectedClientConfig{t, "root"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the stored runtime version of the function for testing so this won't be impacted
			// by other tests having replaced the original with a mock.
			parsePrivateKey = tt.parsePrivateKey
			got, err := runtimeNewClientConfig(tt.args.identityFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("newClientConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (tt.want == nil) != (got == nil) {
				t.Errorf("newClientConfig() = %v, want %v", got, tt.want)
			} else if tt.want != nil {
				tt.want.test(got)
			}
		})
	}
}
