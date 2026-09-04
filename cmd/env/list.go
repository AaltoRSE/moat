/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"sort"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all moat environments",
	Long: `List all moat environments.

This command allows you to view the names of all your current moat environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		// List envs from viper configuration
		viperEnvs := cmd_package.CmdConfig.Sub("envs")
		if viperEnvs == nil {
			fmt.Println("No environments configured.")
			return
		}
		// Unmarshal the environments into a map to get the top-level names
		var envs map[string]types.MoatEnv
		if err := viperEnvs.Unmarshal(&envs); err != nil {
			log.Error().Msgf("Error unmarshalling environments: %v", err)
			return
		}
		// Print the names of the configured environments, sorted for stable output
		names := make([]string, 0, len(envs))
		for name := range envs {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			fmt.Println(name)
		}
	},
}

func init() {
	EnvCmd.AddCommand(listCmd)
}
