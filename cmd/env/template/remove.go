/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/envtemplate"
)

var removeTemplateCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a moat environment template",
	Long: `Remove an existing moat environment template.

This command allows you to delete and remove a previously created environment template.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		if err := envtemplate.RemoveEnvTemplate(cmd_package.CmdConfig, name); err != nil {
			log.Error().Msgf("could not remove environment template: %v", err)
		}
	},
}

func init() {
	var name string

	removeTemplateCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment template to remove")

	if err := removeTemplateCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	templateCmd.AddCommand(removeTemplateCmd)
}
