/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/envtemplate"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

func init() {

	var envName string
	var templateName string
	var homeBase string
	var templateMountString string

	// fromEnvCmd represents the from-env command
	var fromEnvCmd = &cobra.Command{
		Use:   "from-env",
		Short: "Create an environment template from an existing environment",
		Long: `Create an environment template from an existing environment.

This command takes an existing moat environment and converts it into an
environment template, which is then stored in the configuration. The
environment's home directory is used as the template's home base.

The template's home base and mounts can be overwritten with the
--home-base and --mounts flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			overrides := types.EnvTemplate{
				HomeBase: homeBase,
			}
			if templateMountString != "" {
				overrides.Mounts = strings.Split(templateMountString, ",")
			}
			if err := envtemplate.CreateEnvTemplateFromEnv(cmd_package.CmdConfig, envName, templateName, overrides); err != nil {
				log.Error().Msgf("could not create environment template from environment: %v", err)
			}
		},
	}

	fromEnvCmd.Flags().StringVarP(&envName, "env", "e", "", "Name of the existing environment to convert")
	fromEnvCmd.Flags().StringVarP(&templateName, "name", "n", "", "Name of the environment template to create")
	fromEnvCmd.Flags().StringVarP(&homeBase, "home-base", "H", "", "Overwrite the template's home base directory")
	fromEnvCmd.Flags().StringVarP(&templateMountString, "mounts", "m", "", "Overwrite the template's project mounts (comma-separated)")

	if err := fromEnvCmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}
	if err := fromEnvCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	templateCmd.AddCommand(fromEnvCmd)
}
