/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package env

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
	yaml "go.yaml.in/yaml/v3"
)

// CreateEnvShowCmd creates the show subcommand for EnvCmd.
func CreateEnvShowCmd() *cobra.Command {

	// showCmd represents the show command
	var showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of a moat environment",
		Long: `Show the configuration of a moat environment.

This command allows you to view the full configuration of a specific moat environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			env, err := config.GetEnv(config.CmdConfig, name, false)
			if err != nil {
				log.Error().Msgf("Error with the environment %q: %v", name, err)
				return
			}
			env_config, err := yaml.Marshal(env)
			if err != nil {
				log.Error().Msgf("Error marshalling environment: %v", err)
				return
			}
			fmt.Println(string(env_config))
		},
	}

	showCmd.Flags().AddFlagSet(newEnvNameFlagSet())

	if err := showCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return showCmd
}
