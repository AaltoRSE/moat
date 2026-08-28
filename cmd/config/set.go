package cmd

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// setCmd is the cobra command that sets a configuration variable.
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration variable for moat",
	Long: `Set a configuration variable for moat.

This command allows you to set a specific configuration variable for moat.

Examples:
  moat config set defaults.runtime apptainer
  moat config set defaults.runtimes.apptainer.imageurl ghcr.io/aaltorse/vscode-apptainer:latest`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]
		if err := config.SetConfig(key, value); err != nil {
			log.Fatal().Msgf("Failed to set %q: %v", key, err)
		}
		fmt.Printf("Set %s = %s\n", key, value)
	},
}

// init registers the set command with the config command group.
func init() {
	configCmd.AddCommand(setCmd)
}
