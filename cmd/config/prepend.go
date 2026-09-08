package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateConfigPrependCmd creates the prepend subcommand for ConfigCmd.
func CreateConfigPrependCmd() *cobra.Command {

	// prependCmd is the cobra command that prepends a value to a list
	// configuration variable.
	var prependCmd = &cobra.Command{
		Use:   "prepend <key> <value>",
		Short: "Prepend a value to a list configuration variable",
		Long: `Prepend a value to a list configuration variable.

This command prepends a value to an existing list configuration variable.

Examples:
  moat config prepend envs.myenv.mounts /data/project:/data/project
  moat config prepend envs.myenv.readonlymounts /usr/local:/usr/local`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			key, value := args[0], args[1]
			if err := config.PrependConfig(config.CmdConfig, key, value); err != nil {
				log.Fatal().Msgf("Failed to prepend to %q: %v", key, err)
			}
			fmt.Printf("Prepended %s to %s\n", value, key)
		},
	}

	return prependCmd
}
