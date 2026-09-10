package env

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// CreateEnvSetCmd creates the set subcommand for EnvCmd.
func CreateEnvSetCmd() *cobra.Command {

	// setCmd is the cobra command that sets a variable of an existing
	// environment.
	var setCmd = &cobra.Command{
		Use:   "set <key> <value> [<value>]",
		Short: "Set a variable of an existing moat environment",
		Long: `Set a variable of an existing moat environment.

This command allows you to set a specific variable of an existing
environment. Only the environment given by the -n/--name flag is
modified; all other environments and the rest of the configuration
are left untouched.

Examples:
  moat env set -n test command "code --wait ."
  moat env set -n test mountcwd true
  moat env set -n test mounts /path/to/project /run/dbus`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Fatal().Msgf("could not get name flag: %v", err)
			}

			// The environment must already exist
			if _, err := config.GetEnv(config.CmdConfig, name, false); err != nil {
				log.Fatal().Msgf("failed to get environment %q: %v", name, err)
			}

			key := args[0]
			fullKey := "envs." + name + "." + key

			setValue, err := config.SetConfig(config.CmdConfig, fullKey, args[1:])
			if err != nil {
				log.Fatal().Msgf("failed to set %q: %v", fullKey, err)
			}
			fmt.Printf("Set %s.%s = %v\n", name, key, setValue)
		},
	}

	setCmd.Flags().StringP("name", "n", "", "Name of the environment to modify")

	if err := setCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return setCmd
}
