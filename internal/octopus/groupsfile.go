package octopus

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
)

// Allow this to be overridden for tests.
var getAddrsFromGroupsFile = func(hostGroups []string, groupsFile string) ([]string, error) {
	logger.Info.Println("groups file: ", groupsFile)

	f, err := os.Open(groupsFile)
	if err != nil {
		return []string{}, fmt.Errorf("could not load groups file %s: %+v", groupsFile, err)
	}

	fileGroups, err := getAllGroupsInFile(f)
	if err != nil {
		return []string{}, fmt.Errorf("error parsing groups file %s: %+v", groupsFile, err)
	}

	// Make a '${<group>}' argument for each group
	gVars := []string{}
	for _, g := range hostGroups {
		if _, ok := fileGroups[g]; !ok {
			return []string{}, fmt.Errorf("host group %s not found in groups file %s", g, groupsFile)
		}
		gVars = append(gVars, fmt.Sprintf("${%s}", g))
	}

	// Source the hosts file, and echo all the groups without newlines to get all hosts
	cmd := exec.Command("/bin/bash", "-ec",
		fmt.Sprintf("source %s ; echo %s", groupsFile, strings.Join(gVars, " ")))
	o, err := cmd.CombinedOutput()
	// convert to string which has exactly one newline
	os := strings.TrimRight(string(o), "\n")
	if err != nil {
		return []string{}, fmt.Errorf("could not get groups %+v from %s: %+v\n%s", hostGroups, groupsFile, err, os)
	}

	addrs := strings.Split(os, " ")
	return addrs, nil
}

func getAllGroupsInFile(f *os.File) (map[string]bool, error) {
	scanner := bufio.NewScanner(f)
	fileGroups := map[string]bool{}
	// Regex to match Bash variable definition of a host group. Matches: <varname>="
	// <varname> can be any bash variable; the double quote is required
	varRegex, _ := regexp.Compile("^([a-zA-Z_][a-zA-Z0-9_]*)=[\"']")
	for scanner.Scan() {
		l := strings.TrimLeft(scanner.Text(), " \t")
		if m := varRegex.FindStringSubmatch(l); m != nil {
			fileGroups[m[1]] = true
		}
	}
	return fileGroups, scanner.Err()
}
