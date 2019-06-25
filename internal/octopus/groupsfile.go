package octopus

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/BlaineEXE/octopus/internal/logger"
)

// Allow this to be overridden for tests.
var getAddrsFromGroupsFile = func(hostGroups []string, groupsFile string) ([]string, error) {
	logger.Info.Println("groups file: ", groupsFile)

	fileGroups, err := getAllGroupsInFile(groupsFile)
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
	cmd := exec.Command("bash", "-ec",
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

func getAllGroupsInFile(filePath string) (map[string]bool, error) {
	errMsg := "failed to parse groups from groups file"
	retGroups := map[string]bool{}

	// get the set of environment variables which bash reports by default, e.g., PWD, SHLVL, _
	// env --ignore-environment is not available on mac, so must use "-i"
	getBaseEnv := exec.Command("env", "-i", "-", "bash", "-c", "env")
	baseEnv, err := getBaseEnv.Output()
	if err != nil {
		return retGroups, fmt.Errorf("%s. failed to determine base environment variables. %+v", errMsg, err)
	}
	baseVars := map[string]bool{}
	b := bytes.NewReader(baseEnv)
	scanner := bufio.NewScanner(b)
	for scanner.Scan() {
		v := parseLine(scanner.Text())
		if v != "" {
			baseVars[v] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return retGroups, fmt.Errorf("%s. failed to parse base environment variables. %+v", errMsg, err)
	}

	// get the set of environment variables which bash reports after loading the group file
	// this will also report the base env vars found above, which we want to ignore
	fileEnvCmd := fmt.Sprintf("source %s && env", filePath)
	getFileEnv := exec.Command("env", "-i", "-", "bash", "-c", fileEnvCmd)
	fileEnv, err := getFileEnv.Output()
	if err != nil {
		return retGroups, fmt.Errorf("%s. failed to determine exported host groups file variables. %+v", errMsg, err)
	}

	b = bytes.NewReader(fileEnv)
	scanner = bufio.NewScanner(b)
	for scanner.Scan() {
		l := scanner.Text()
		logger.Info.Println(l)
		v := parseLine(l)
		if v == "" {
			continue
		}
		if _, ok := baseVars[v]; ok {
			// if env is in base envs, don't count it as a group
			continue
		}
		retGroups[v] = true
	}
	if err := scanner.Err(); err != nil {
		return retGroups, fmt.Errorf("%s. failed to parse exported host groups file variables. %+v", errMsg, err)
	}

	return retGroups, nil
}

// parse a bash variable name from a line returned by `env`
// variable names match regex [a-zA-Z_][a-zA-Z0-9_]* and always end with an equal sign
// variables may have multiline strings which will start on the line following the var name, so
//   not all lines will have a variable
func parseLine(line string) string {
	if len(line) == 0 {
		return ""
	}
	c := line[0]
	if (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) && c != '_' {
		return ""
	}
	v := make([]byte, 0, len(line))
	v = append(v, c)
	for i := 1; i < len(line); i++ {
		c = line[i]
		if c == '=' {
			break
		}
		if (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) && c != '_' && (c < '0' || c > '9') {
			return ""
		}
		v = append(v, c)
	}
	return string(v)
}
