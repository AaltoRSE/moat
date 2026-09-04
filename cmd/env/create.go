/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

var name string
var home string
var mountString string
var yes bool

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new moat environment",
	Long: `Create a new moat environment.

This command allows you to create and configure a new environment.
Use --yes to create missing directories (fake home and mount paths)
without prompting.`,
	Run: func(cmd *cobra.Command, args []string) {
		var mounts []string
		if mountString != "" {
			mounts = strings.Split(mountString, ",")
		}
		if err := env.CreateEnvironment(cmd_package.CmdConfig, name, types.MoatEnv{
			Home:           home,
			Mounts:         mounts,
			ReadOnlyMounts: []string{},
		}, yes); err != nil {
			log.Error().Msgf("could not create environment: %v", err)
		}
	},
}

func init() {
	createCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment")
	createCmd.Flags().StringVarP(&home, "home", "H", "", "Home directory for the environment")
	createCmd.Flags().StringVarP(&mountString, "mounts", "m", "", "Comma-separated list of project mounts")
	createCmd.Flags().BoolVarP(&yes, "yes", "y", false, "Create missing directories (fake home and mount paths) without prompting")

	if err := createCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := createCmd.MarkFlagRequired("home"); err != nil {
		panic(err)
	}
	EnvCmd.AddCommand(createCmd)
}
