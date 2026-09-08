package config

import (
	"fmt"
	"reflect"

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

			var value any

			key := args[0]

			// Check key type
			keyType, err := config.GetVariableType(config.CmdConfig, key)
			if err != nil {
				log.Fatal().Msgf("Failed to get type for key %q: %v", key, err)
			}
			if keyType.Kind() == reflect.String {
				// Return error if there are more than 2 arguments for a string key
				if len(args) > 2 {
					log.Fatal().Msgf("Too many arguments for key %q of type string", key)
				}
				value = args[1]
			} else if keyType.Kind() == reflect.Slice {
				value = args[1:]
			} else {
				log.Fatal().Msgf("Unsupported key type for key %q: %v", key, keyType.Kind())
			}

			if err := config.SetConfig(config.CmdConfig, key, value); err != nil {
				log.Fatal().Msgf("Failed to set %q: %v", key, err)
			}
			fmt.Printf("Set %s = %s\n", key, value)
		},
	}

	return setCmd
}
