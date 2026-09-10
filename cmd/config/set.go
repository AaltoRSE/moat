package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateConfigSetCmd creates the set subcommand for ConfigCmd.
func CreateConfigSetCmd() *cobra.Command {

	// setCmd is the cobra command that sets a configuration variable.
	var setCmd = &cobra.Command{
		Use:   "set <key> <value> [<value>]",
		Short: "Set a configuration variables for moat",
		Long: `Set a configuration variable for moat.

This command allows you to set a specific configuration variable for moat.

Examples:
  moat config set defaults.runtimes.apptainer.imageurl ghcr.io/aaltorse/vscode-apptainer:latest`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			key := args[0]

			setValue, err := config.SetConfig(config.CmdConfig, key, args[1:])
			if err != nil {
				log.Fatal().Msgf("failed to set %q: %v", key, err)
			}
			fmt.Printf("Set %s = %v\n", key, setValue)
		},
	}

	return setCmd
}
