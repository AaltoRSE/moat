package env

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// CreateEnvSetCmd creates the set subcommand for EnvCmd.
func CreateEnvSetCmd() *cobra.Command {

	// update pairs an environment variable name with the value to set it to.
	type update struct {
		key   string
		value []string
	}

	// setCmd is the cobra command that sets variables of an existing
	// environment.
	var setCmd = &cobra.Command{
		Use:   "set",
		Short: "Set a variable of an existing moat environment",
		Long: `Set a variable of an existing moat environment.

This command allows you to update the variables of an existing
environment. Only the environment given by the -n/--name flag is
modified; all other environments and the rest of the configuration
are left untouched. Only the flags that are given are changed; all
other variables of the environment keep their current values.

Examples:
  moat env set -n test --command "code --wait ."
  moat env set -n test --home /home/user/project
  moat env set -n test --mount /path/to/project --ro-mount /run/dbus`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Fatal().Msgf("could not get name flag: %v", err)
			}

			// The environment must already exist
			if _, err := config.GetEnv(config.CmdConfig, name, false); err != nil {
				log.Fatal().Msgf("failed to get environment %q: %v", name, err)
			}

			// Collect the flags that were explicitly given, in a fixed
			// order, so that the output is deterministic.
			var updates []update
			if cmd.Flags().Changed("home") {
				home, err := cmd.Flags().GetString("home")
				if err != nil {
					log.Fatal().Msgf("could not get home flag: %v", err)
				}
				updates = append(updates, update{"home", []string{home}})
			}
			if cmd.Flags().Changed("mount") {
				mounts, err := cmd.Flags().GetStringArray("mount")
				if err != nil {
					log.Fatal().Msgf("could not get mount flag: %v", err)
				}
				updates = append(updates, update{"mounts", mounts})
			}
			if cmd.Flags().Changed("ro-mount") {
				roMounts, err := cmd.Flags().GetStringArray("ro-mount")
				if err != nil {
					log.Fatal().Msgf("could not get ro-mount flag: %v", err)
				}
				updates = append(updates, update{"readonlymounts", roMounts})
			}
			if cmd.Flags().Changed("command") {
				command, err := cmd.Flags().GetString("command")
				if err != nil {
					log.Fatal().Msgf("could not get command flag: %v", err)
				}
				updates = append(updates, update{"command", []string{command}})
			}

			if len(updates) == 0 {
				log.Fatal().Msgf("no variables to set: give at least one of --home, --mount, --ro-mount or --command")
			}

			for _, u := range updates {
				setValue, err := config.SetConfig(config.CmdConfig, "envs."+name+"."+u.key, u.value)
				if err != nil {
					log.Fatal().Msgf("failed to set %q: %v", u.key, err)
				}
				fmt.Printf("Set %s.%s = %v\n", name, u.key, setValue)
			}
		},
	}

	setCmd.Flags().StringP("name", "n", "", "Name of the environment to modify")
	setCmd.Flags().StringP("home", "H", "", "New home directory for the environment")
	setCmd.Flags().StringArrayP("mount", "m", nil, "Mounted directory (can be specified multiple times, replaces all existing mounts)")
	setCmd.Flags().StringArrayP("ro-mount", "r", nil, "Read-only mounted directory (can be specified multiple times, replaces all existing read-only mounts)")
	setCmd.Flags().StringP("command", "C", "", "Command to run in the environment (use quotes for multi-word commands)")

	if err := setCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return setCmd
}
