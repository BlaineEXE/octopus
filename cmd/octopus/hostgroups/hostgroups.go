// Package hostgroups defines the 'octopus host-groups' command behavior. It's one responsibility
// is to report the list of host groups octopus is able to parse from the host groups file.
package hostgroups

import (
	"fmt"
	"log"

	"github.com/BlaineEXE/octopus/internal/octopus"
	"github.com/BlaineEXE/octopus/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	aboutText = "Parse the host groups file, and report the available groups"
)

// HostGroupsCommand is the 'host-groups' command definition which prints the valid host groups.
var HostGroupsCommand = &cobra.Command{
	Use:   "host-groups",
	Short: aboutText,
	Long:  fmt.Sprintf("\n%s", aboutText),
	Args:  cobra.ExactArgs(0), // support no args
	RunE: func(cmd *cobra.Command, args []string) error {
		o := octopus.New(
			nil,
			[]string{},
			getAbsFilePath(viper.GetString("groups-file")),
		)

		gs, err := o.ValidHostGroups()
		if err != nil {
			return err
		}
		for _, g := range gs {
			fmt.Printf("%s\n", g)
		}
		return nil
	},
}

func getAbsFilePath(path string) string {
	a, err := util.AbsPath(path)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return a
}
