/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"

	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
	yaml "go.yaml.in/yaml/v3"
)

func init() {

	var name string

	// showCmd represents the show command
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show the configuration of a moat environment",
		Long: `Show the configuration of a moat environment.

This command allows you to view the full configuration of a specific moat environment.`,
		Run: func(cmd *cobra.Command, args []string) {
			env, err := config.GetEnv(cmd_package.CmdConfig, name, false)
			if err != nil {
				log.Error().Msgf("Error with the environment %q: %v", name, err)
				return
			}
			bs, err := yaml.Marshal(env)
			if err != nil {
				log.Error().Msgf("Error marshalling environment: %v", err)
				return
			}
			fmt.Println(string(bs))
		},
	}

	showCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the environment to show")

	if err := showCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	EnvCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
