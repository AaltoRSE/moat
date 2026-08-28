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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all moat environments",
	Long: `List all moat environments.

This command allows you to view all your current moat environments in a structured format.`,
	Run: func(cmd *cobra.Command, args []string) {
		// List envs from viper configuration
		viperEnvs := viper.Sub("envs")
		// Unmarshalling the environments into a map for easier iteration
		if viperEnvs != nil {
			c := viperEnvs.AllSettings()
			bs, err := yaml.Marshal(c)
			if err != nil {
				log.Print("Error marshalling environments:", err)
				return
			}
			fmt.Println(string(bs))
		} else {
			fmt.Println("No environments configured.")
		}
	},
}

func init() {
	EnvCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
