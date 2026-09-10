package env

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/spf13/cobra"
)

// CreateEnvRemoveCmd creates the remove subcommand for EnvCmd.
func CreateEnvRemoveCmd() *cobra.Command {

	// removeCmd represents the remove command
	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove an environment from moat",
		Long: `Remove an existing moat environment.

This command allows you to delete and remove a previously created environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			if err := env.RemoveEnvironment(config.CmdConfig, name); err != nil {
				log.Error().Msgf("could not remove environment: %v", err)
			}
		},
	}

	removeCmd.Flags().AddFlagSet(newEnvNameFlagSet())

	if err := removeCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return removeCmd
}
