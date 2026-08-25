/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration variable for shark-tank",
	Long: `Set a configuration variable for shark-tank.

This command allows you to set a specific configuration variable for shark-tank.

Examples:
  shark-tank config set defaults.runtime apptainer
  shark-tank config set defaults.runtimes.apptainer.imageurl ghcr.io/aaltorse/vscode-apptainer:latest`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key, value := args[0], args[1]
		if err := config.SetConfig(key, value); err != nil {
			log.Fatal().Msgf("Failed to set %q: %v", key, err)
		}
		fmt.Printf("Set %s = %s\n", key, value)
	},
}

func init() {
	configCmd.AddCommand(setCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
