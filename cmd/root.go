/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.shark-tank.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

	// In case of error, print the configuration
	if err := config.InitConfig(); err != nil {
		log.Fatalf("could not initialize config: %v", err)
	}
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
