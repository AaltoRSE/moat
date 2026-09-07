/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/spf13/cobra"
)

func init() {

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
			command, err := cmd.Flags().GetStringArray("command")
			if err != nil {
				log.Error().Msgf("could not get command flag: %v", err)
				return
			}
			// Sanitize the command, discarding any environment variables.
			_, command, err = utils.SanitizeArgs(command)
			if err != nil {
				log.Error().Msgf("could not sanitize command: %v", err)
				return
			}
			yes, err := cmd.Flags().GetBool("yes")
			if err != nil {
				log.Error().Msgf("could not get yes flag: %v", err)
				return
			}

			if err := env.CopyEnvironment(cmd_package.CmdConfig, source, name, home, mounts, command, yes); err != nil {
				log.Error().Msgf("could not copy environment: %v", err)
			}
		},
	}

	copyCmd.Flags().StringP("source", "s", "", "Name of the source environment to copy from")
	copyCmd.Flags().StringP("name", "n", "", "Name of the new environment to create")
	copyCmd.Flags().StringP("home", "H", "", "Home directory for the new environment (overrides the source)")
	copyCmd.Flags().StringArrayP("mounts", "m", nil, "Array of project mounts (overrides the source)")
	copyCmd.Flags().StringArrayP("command", "C", nil, "Array of commands to run in the new environment (overrides the source)")
	copyCmd.Flags().BoolP("yes", "y", false, "Create missing directories (fake home and mount paths) without prompting")

	if err := copyCmd.MarkFlagRequired("source"); err != nil {
		panic(err)
	}
	if err := copyCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	EnvCmd.AddCommand(copyCmd)
}
