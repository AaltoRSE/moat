package cmd

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var appendCmd = &cobra.Command{
	Use:   "append <key> <value>",
	Short: "Append a value to a list configuration variable",
	Long: `Append a value to a list configuration variable.

This command appends a value to an existing list configuration variable.

Examples:
  moat config append envs.myenv.mounts /data/project:/data/project
  moat config append envs.myenv.readonlymounts /usr/local:/usr/local`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]
		if err := config.AppendConfig(key, value); err != nil {
			log.Fatal().Msgf("Failed to append to %q: %v", key, err)
		}
		fmt.Printf("Appended %s to %s\n", value, key)
	},
}

func init() {
	configCmd.AddCommand(appendCmd)
}
