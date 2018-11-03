package octopus

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"testing"
)

var runtimeGetAddrsFromGroupsFile func(hostGroups []string, groupsFile string) ([]string, error)

func init() {
	// On init, store the default version of getAddrsFromGroupsFile which will execute at runtime
	// so it isn't lost by tests replacing this with mocks.
	runtimeGetAddrsFromGroupsFile = getAddrsFromGroupsFile
}

const parsableGroups = `
# simple
a="1.1.1.1"
b='2.2.2.2'
_3="3.3.3.3"
_4='4.4.4.4'

# double-quoted multi, all variants
d34="3.3.3.3 4.4.4.4"
d5_6="5.5.5.5
6.6.6.6"
d_78="
7.7.7.7 8.8.8.8"
d_9_10="
9.9.9.9
10.10.10.10"
d_11_12_="
11.11.11.11
12.12.12.12
"

# single-quoted multi
s56='5.5.5.5 6.6.6.6'
s_7_8_='
7.7.7.7
8.8.8.8
'

# leading/trailing space
ltd78=" 7.7.7.7 8.8.8.8 "
lts910=' 9.9.9.9 10.10.10.10 '

# mixed
md9to12="9.9.9.9 10.10.10.10
11.11.11.11 12.12.12.12"
ms13to16='13.13.13.13 14.14.14.14
15.15.15.15 16.16.16.16'
`

const noGroups = `
#!/usr/bin/env bash
`

const unparsableGroups = `
a='9.9.9.9'
t="1.1.1.1'
`

func createFile(t *testing.T, path, contents string, writeonly bool) {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		t.Fatalf("Could not create test file: %s", path)
	}
	if _, err := f.WriteString(contents); err != nil {
		t.Fatalf("Could not write to test file: %s", path)
	}
	if writeonly { // write only files will prevent file from being read, should throw error
		os.Chmod(path, 0222)
	}
}

func Test_getAddrsFromGroupsFile(t *testing.T) {
	// Make temp dir for testing
	tmpRoot, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("failed to create temp dir for testing. %+v", err)
	}
	defer os.RemoveAll(tmpRoot)
	goodGroupsFile := path.Join(tmpRoot, "goodGroups")
	noGroupsFile := path.Join(tmpRoot, "noGroups")
	unparsableGroupsFile := path.Join(tmpRoot, "unparsableGroups")
	writeonlyGroupsFile := path.Join(tmpRoot, "writeonlyGroupsFile")
	createFile(t, goodGroupsFile, parsableGroups, false)
	createFile(t, noGroupsFile, noGroups, false)
	createFile(t, unparsableGroupsFile, unparsableGroups, false)
	createFile(t, writeonlyGroupsFile, parsableGroups, true)
	b, _ := ioutil.ReadFile(goodGroupsFile)
	fmt.Println(string(b))

	tests := []struct {
		name       string
		groupsFile string
		hostGroups []string
		want       []string
		wantErr    bool
	}{
		{"unreadable file", writeonlyGroupsFile, []string{"a"}, []string{}, true},
		{"no groups in file", noGroupsFile, []string{"a"}, []string{}, true},
		{"unparsable file", unparsableGroupsFile, []string{"a", "t"}, []string{}, true},
		{"group not in file", goodGroupsFile, []string{"a", "notIn"}, []string{}, true},
		{"simple double", goodGroupsFile, []string{"a", "_3"}, []string{"1.1.1.1", "3.3.3.3"}, false},
		{"simple single", goodGroupsFile, []string{"b", "_4"}, []string{"2.2.2.2", "4.4.4.4"}, false},
		{"double multi", goodGroupsFile, []string{"d34", "d5_6", "d_78", "d_9_10", "d_11_12_"},
			[]string{"3.3.3.3", "4.4.4.4", "5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8", "9.9.9.9",
				"10.10.10.10", "11.11.11.11", "12.12.12.12"}, false},
		{"single multi", goodGroupsFile, []string{"s56", "s_7_8_"},
			[]string{"5.5.5.5", "6.6.6.6", "7.7.7.7", "8.8.8.8"}, false},
		{"leading+trailing space", goodGroupsFile, []string{"ltd78", "lts910"},
			[]string{"7.7.7.7", "8.8.8.8", "9.9.9.9", "10.10.10.10"}, false},
		{"mixed", goodGroupsFile, []string{"md9to12", "ms13to16"},
			[]string{"9.9.9.9", "10.10.10.10", "11.11.11.11", "12.12.12.12",
				"13.13.13.13", "14.14.14.14", "15.15.15.15", "16.16.16.16"}, false},
		{"arbitrary, out of order selection", goodGroupsFile, []string{"md9to12", "ltd78", "_4"},
			[]string{"9.9.9.9", "10.10.10.10", "11.11.11.11", "12.12.12.12",
				"7.7.7.7", "8.8.8.8", "4.4.4.4"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the stored runtime version of the function for testing so this won't be impacted
			// by other tests having replaced the original with a mock.
			got, err := runtimeGetAddrsFromGroupsFile(tt.hostGroups, tt.groupsFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAddrsFromGroupsFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAddrsFromGroupsFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
