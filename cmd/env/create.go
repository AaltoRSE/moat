package env

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

func CreateEnvCreateCmd() *cobra.Command {

	// createCmd represents the create command
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new moat environment",
		Long: `Create a new moat environment.

This command allows you to create and configure a new environment.
Use --yes to create missing directories (fake home and mount paths)
without prompting.`,
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				log.Error().Msgf("could not get yes flag: %v", err)
				return
			}
			home, err := cmd.Flags().GetString("home")
			if err != nil {
				log.Error().Msgf("could not get home flag: %v", err)
				return
			}
			mounts, err := cmd.Flags().GetStringArray("mounts")
			if err != nil {
				log.Error().Msgf("could not get mounts flag: %v", err)
				return
			}
			roMounts, err := cmd.Flags().GetStringArray("ro-mounts")
			if err != nil {
				log.Error().Msgf("could not get ro-mounts flag: %v", err)
				return
			}
			command, err := cmd.Flags().GetString("command")
			if err != nil {
				log.Error().Msgf("could not get command flag: %v", err)
				return
			}
			var commandPtr *string
			if command != "" {
				commandPtr = &command
			}
			moatEnv := types.MoatEnv{
				Home:           home,
				Mounts:         mounts,
				ReadOnlyMounts: roMounts,
				Command:        commandPtr,
			}

			if err := env.CreateEnvironment(config.CmdConfig, name, moatEnv, yes); err != nil {
				log.Error().Msgf("could not create environment: %v", err)
			}
		},
	}

	createCmd.Flags().StringP("name", "n", "", "Name of the environment")
	createCmd.Flags().StringP("home", "H", "", "Home directory for the environment")
	createCmd.Flags().StringArrayP("mounts", "m", nil, "Array of project mounts")
	createCmd.Flags().StringArrayP("ro-mounts", "r", nil, "Array of read-only project mounts")
	createCmd.Flags().StringP("command", "C", "", "Array of commands to run in the environment")
	createCmd.Flags().BoolP("yes", "y", false, "Create missing directories (fake home and mount paths) without prompting")

	if err := createCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := createCmd.MarkFlagRequired("home"); err != nil {
		panic(err)
	}

	return createCmd
}
