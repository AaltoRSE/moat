/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/spf13/cobra"
)

func init() {

	var source string
	var name string
	var home string
	var mountString string
	var yes bool

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
			var mounts []string
			if mountString != "" {
				mounts = strings.Split(mountString, ",")
			}
			if err := env.CopyEnvironment(cmd_package.CmdConfig, source, name, home, mounts, yes); err != nil {
				log.Error().Msgf("could not copy environment: %v", err)
			}
		},
	}

	copyCmd.Flags().StringVarP(&source, "source", "s", "", "Name of the source environment to copy from")
	copyCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the new environment to create")
	copyCmd.Flags().StringVarP(&home, "home", "H", "", "Home directory for the new environment (overrides the source)")
	copyCmd.Flags().StringVarP(&mountString, "mounts", "m", "", "Comma-separated list of project mounts (overrides the source)")
	copyCmd.Flags().BoolVarP(&yes, "yes", "y", false, "Create missing directories (fake home and mount paths) without prompting")

	if err := copyCmd.MarkFlagRequired("source"); err != nil {
		panic(err)
	}
	if err := copyCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	EnvCmd.AddCommand(copyCmd)
}
