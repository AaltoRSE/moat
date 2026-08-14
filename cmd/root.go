/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/rs/zerolog/log"

	config "github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "shark-tank",
	Short: "A brief description of your application",
	Long: `Shark-tank is an application for running AI agents in
containerized environtment.`,
	DisableFlagParsing: false,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

}

func initConfig() {
	configFile, _ := RootCmd.Flags().GetString("config")

	if err := config.InitConfig(configFile); err != nil {
		log.Error().Err(err).Msgf("could not initialize config: %v", err)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	var cfgFile string

	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/shark-tank/config.yaml or config.yaml in the current directory)")

	log.Debug().Msgf("Using config file: %s", cfgFile)
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
