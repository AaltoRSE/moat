/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/spf13/cobra"
)

func init() {

	var name string

	// removeCmd represents the remove command
	var removeCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove an environment from moat",
		Long: `Remove an existing moat environment.

This command allows you to delete and remove a previously created environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			log.Print("remove called")
			if err := env.RemoveEnvironment(cmd_package.CmdConfig, name); err != nil {
				log.Printf("Error removing environment: %v\n", err)
				return
			}
			log.Print("Environment successfully removed.\n")
		},
	}

	removeCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment to remove")

	if err := removeCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	EnvCmd.AddCommand(removeCmd)
}
