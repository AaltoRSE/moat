/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/shark-tank/internal/envtemplate"
	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/cobra"
)

var templateName string
var homeBase string
var templateMountString string

var createTemplateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new shark-tank environment template",
	Long: `Create a new shark-tank environment template.

This command allows you to create and configure a new environment template.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := envtemplate.CreateEnvTemplate(templateName, types.EnvTemplate{
			HomeBase:       homeBase,
			Mounts:         strings.Split(templateMountString, ","),
			ReadOnlyMounts: []string{},
		}); err != nil {
			log.Error().Msgf("could not create environment template: %v", err)
		}
	},
}

func init() {
	createTemplateCmd.Flags().StringVarP(&templateName, "name", "n", "", "Name of the environment template")
	createTemplateCmd.Flags().StringVarP(&homeBase, "home-base", "H", "", "Base directory for environment homes")
	createTemplateCmd.Flags().StringVarP(&templateMountString, "mounts", "m", "", "Comma-separated list of project mounts")

	if err := createTemplateCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := createTemplateCmd.MarkFlagRequired("home-base"); err != nil {
		panic(err)
	}

	templateCmd.AddCommand(createTemplateCmd)
}
