package config

import (
	"os"

	"github.com/BlaineEXE/octopus/internal/logger"
	"github.com/BlaineEXE/octopus/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultGroupsFile = "_node-list"
)

// OctopusCmd is the top-level 'octopus' command.
var OctopusCmd = &cobra.Command{
	Use:   "octopus [flags] [--host-groups|-h <HOST-GROUPS>] <COMMAND>",
	Short: "Octopus runs a command on multiple remote hosts in parallel",
	Long: `
-----------
  OCTOPUS
-----------

  Octopus is a simple pdsh-inspired commandline tool for running the same
  command on multiple remote hosts in parallel. Hosts are grouped together
  into "host groups" in a file which inspired by a "genders" file. The
  host groups file for Octopus is actually a Bash file with groups defined by
  variable definitions. This is so that the same file may be used easily by
	both Octopus and by user-made scripts and has the secondary benefit of
	supporting defining hosts by IP address as well as hostname.

  Under the hood, Octopus uses ssh connections, and some ssh arguments are
  reflected in Octopus's arguments. These arguments are marked in the help
  text with "(ssh)".

  WARNINGS:
    Octopus does not do verification of remote hosts (Equivalent to
    setting ssh option StrictHostKeyChecking=no) and does not add entries
    to the known hosts file.

  Config file:
    Octopus supports setting custom default values for flags in a config file.
  Any of Octopus's top-level, full-name flags can be set in the config.
  Octopus will search in order for the first 'config.yaml' file it finds in:
    (1) ./.octopus
    (2) $HOME/.octopus
    (3) /etc/octopus
  e.g., Simply writing "host-groups: all" into the config file will use the
  'all' host group for Octopus commands unless the user specifies a different
  set of host groups using '--host-groups|-h' on the commandline.
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("verbose") {
			logger.Info.SetOutput(os.Stderr)
			logger.Info.Println("Running octopus in verbose mode")
			logger.Info.Println("Octopus version:", version.Version)
		}
	},
	BashCompletionFunction: bashCompletionFunc,
}

// SetCmdFlagCompletion sets a custom completion function for a flag.
func SetCmdFlagCompletion(cmd *cobra.Command, flag, completionFunction string) {
	if cmd.Flag(flag).Annotations == nil {
		cmd.Flag(flag).Annotations = map[string][]string{}
	}
	f := cmd.Flag(flag)
	f.Annotations[cobra.BashCompCustom] = append(
		f.Annotations[cobra.BashCompCustom],
		completionFunction,
	)
}

// BashCompletionEmptyCompletionFunction is the name of the custom completion function which will
// return empty completion for a flag. Use this when Bash's default behavior of suggesting files in
// the current directory aren't useful for a flag.
const BashCompletionEmptyCompletionFunction = "__octopus_empty_completion"

const bashCompletionFunc = `
__octopus_empty_completion()
{
	# suggest 2 different, empty completions to stop bash from returning default completions to user
	COMPREPLY+=("" " ")
}

# taken from kubectl; override flags are flags that need to be present when running octopus to get
# completion suggestions for other flags because they change the behavior of octopus.
# specifically, groups-file changes the groups available to octopus for getting completion for the
# host-groups flag
__octopus_override_flag_list=(--groups-file -f)
__octopus_override_flags()
{
    local ${__octopus_override_flag_list[*]##*-} two_word_of of var
    for w in "${words[@]}"; do
        if [ -n "${two_word_of}" ]; then
            eval "${two_word_of##*-}=\"${two_word_of}=\${w}\""
            two_word_of=
            continue
        fi
        for of in "${__octopus_override_flag_list[@]}"; do
            case "${w}" in
                ${of}=*)
                    eval "${of##*-}=\"${w}\""
                    ;;
                ${of})
                    two_word_of="${of}"
                    ;;
            esac
        done
    done
    for var in "${__octopus_override_flag_list[@]##*-}"; do
        if eval "test -n \"\$${var}\""; then
            eval "echo \${${var}}"
        fi
    done
}

__octopus_get_host_groups()
{
	local out
	if out=$(octopus $(__octopus_override_flags) host-groups); then
		if [[ -n "$out" ]]; then
			COMPREPLY+=( $( compgen -W "${out[*]}" -- "$cur" ) )
		else
			__octopus_empty_completion
		fi
	fi
}
`

func init() {
	// Load the config file at cobra initialization
	cobra.OnInitialize(loadConfig)

	// Persistent top-level flags
	OctopusCmd.PersistentFlags().StringP("groups-file", "f", defaultGroupsFile,
		"file which defines groups of remote hosts available for execution")

	OctopusCmd.PersistentFlags().StringSliceP("host-groups", "g", []string{},
		"comma-separated list of host groups; the command will be run on each host in every group")
	SetCmdFlagCompletion(OctopusCmd, "host-groups", "__octopus_get_host_groups")

	OctopusCmd.PersistentFlags().StringP("identity-file", "i", "$HOME/.ssh/id_rsa",
		"(ssh) file from which the identity (private key) for public key authentication is read")

	OctopusCmd.PersistentFlags().Uint16P("port", "p", 22,
		"(ssh) port on which to connect to hosts")
	SetCmdFlagCompletion(OctopusCmd, "port", BashCompletionEmptyCompletionFunction)

	OctopusCmd.PersistentFlags().StringP("user", "u", "root",
		"user as which to connect to hosts (corresponds to ssh \"-l\" option)")
	SetCmdFlagCompletion(OctopusCmd, "user", BashCompletionEmptyCompletionFunction)

	OctopusCmd.PersistentFlags().BoolP("verbose", "v", false,
		"print additional information about octopus progress")

	// Persistent flags are also valid config file options
	viper.BindPFlags(OctopusCmd.PersistentFlags())
}
