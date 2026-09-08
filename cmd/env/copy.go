package env

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/spf13/cobra"
)

// CreateEnvCopyCmd creates the copy subcommand for EnvCmd.
func CreateEnvCopyCmd() *cobra.Command {

	// copyCmd represents the copy command
	copyCmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy an existing moat environment",
		Long: `Copy an existing moat environment.

This command creates a new environment as a copy of an existing one.
The new environment inherits all settings from the source environment,
and you may optionally override the home directory and project mounts.
Use --yes to create missing directories (fake home and mount paths)
without prompting.`,
		Run: func(cmd *cobra.Command, args []string) {
			source, err := cmd.Flags().GetString("source")
			if err != nil {
				log.Error().Msgf("could not get source flag: %v", err)
				return
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
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
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				log.Error().Msgf("could not get yes flag: %v", err)
				return
			}

			if err := env.CopyEnvironment(config.CmdConfig, source, name, home, mounts, roMounts, commandPtr, yes); err != nil {
				log.Error().Msgf("could not copy environment: %v", err)
			}
		},
	}

	copyCmd.Flags().StringP("source", "s", "", "Name of the source environment to copy from")
	copyCmd.Flags().StringP("name", "n", "", "Name of the new environment to create")
	copyCmd.Flags().StringP("home", "H", "", "Home directory for the new environment (overrides the source)")
	copyCmd.Flags().StringArrayP("mounts", "m", nil, "Array of project mounts (overrides the source)")
	copyCmd.Flags().StringArrayP("ro-mounts", "r", nil, "Array of read-only project mounts (overrides the source)")
	copyCmd.Flags().StringP("command", "C", "", "Command to run in the new environment (overrides the source)")
	copyCmd.Flags().BoolP("yes", "y", false, "Create missing directories (fake home and mount paths) without prompting")

	if err := copyCmd.MarkFlagRequired("source"); err != nil {
		panic(err)
	}
	if err := copyCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return copyCmd
}
