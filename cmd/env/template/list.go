/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

var listTemplateCmd = &cobra.Command{
	Use:   "list",
	Short: "List all moat environment templates",
	Long: `List all moat environment templates.

This command allows you to view all your current moat environment templates in a structured format.`,
	Run: func(cmd *cobra.Command, args []string) {
		viperTemplates := viper.Sub("envtemplates")
		if viperTemplates != nil {
			c := viperTemplates.AllSettings()
			bs, err := yaml.Marshal(c)
			if err != nil {
				log.Error().Err(err).Msg("Error marshalling environment templates")
				return
			}
			fmt.Println(string(bs))
		} else {
			fmt.Println("No environment templates configured.")
		}
	},
}

func init() {
	templateCmd.AddCommand(listTemplateCmd)
}
