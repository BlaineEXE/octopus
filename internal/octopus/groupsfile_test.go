package octopus

import (
	"path"
	"reflect"
	"testing"

	"github.com/BlaineEXE/octopus/internal/util/testutil"
)

var runtimeGetAddrsFromGroupsFile func(hostGroups []string, groupsFile string) ([]string, error)

func init() {
	// On init, store the default version of getAddrsFromGroupsFile which will execute at runtime
	// so it isn't lost by tests replacing this with mocks.
	runtimeGetAddrsFromGroupsFile = getAddrsFromGroupsFile
}

const parsableGroups = `
# simple
export a="1.1.1.1"
export b='2.2.2.2'
export _3="3.3.3.3"
export _4='4.4.4.4'

# double-quoted multi, all variants
export d34="3.3.3.3 4.4.4.4"
export d5_6="5.5.5.5
6.6.6.6"
export d_78="
7.7.7.7 8.8.8.8"
export d_9_10="
9.9.9.9
10.10.10.10"
export d_11_12_="
11.11.11.11
12.12.12.12
"

# single-quoted multi
export s56='5.5.5.5 6.6.6.6'
export s_7_8_='
7.7.7.7
8.8.8.8
'

# leading/trailing space
export ltd78=" 7.7.7.7 8.8.8.8 "
export lts910=' 9.9.9.9 10.10.10.10 '

# mixed
export md9to12="9.9.9.9 10.10.10.10
11.11.11.11 12.12.12.12"
export ms13to16='13.13.13.13 14.14.14.14
15.15.15.15 16.16.16.16'
`

const noGroups = `
#!/usr/bin/env bash
`

const unparsableGroups = `
export a='9.9.9.9'
export t="1.1.1.1'
`

const invalidVarName = `
export a&b='1.1.1.1'
`

func Test_getAddrsFromGroupsFile(t *testing.T) {
	// Make temp dir for testing
	tmpRoot, cleanup := testutil.TempDir("")
	defer cleanup()
	goodGroupsFile := path.Join(tmpRoot, "goodGroups")
	noGroupsFile := path.Join(tmpRoot, "noGroups")
	unparsableGroupsFile := path.Join(tmpRoot, "unparsableGroups")
	writeonlyGroupsFile := path.Join(tmpRoot, "writeonlyGroupsFile")
	testutil.WriteFile(goodGroupsFile, parsableGroups, 0644)
	testutil.WriteFile(noGroupsFile, noGroups, 0644)
	testutil.WriteFile(unparsableGroupsFile, unparsableGroups, 0644)
	testutil.WriteFile(writeonlyGroupsFile, parsableGroups, 0222)

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
		{"invalid var name", invalidVarName, []string{}, []string{}, true},
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
