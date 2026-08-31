/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	config "github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "moat",
	Short: "A brief description of your application",
	Long: `Moat is an application for running AI agents in
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
	log.Debug().Interface("configFile", configFile).Msg("Using config file: ")

	if err := config.InitConfig(configFile); err != nil {
		log.Error().Err(err).Msgf("could not initialize config: %v", err)
	}
}

func initLogger() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Default level for this example is info, unless debug flag is present
	debug, _ := RootCmd.Flags().GetBool("debug")
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Debug().Msgf("Logger initialized. Debug mode: %v", zerolog.GlobalLevel() == zerolog.DebugLevel)
}

func initServices() {
	initLogger()
	initConfig()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	var debug bool
	var cfgFile string

	debug = false
	cfgFile = ""

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/moat/config.yaml or config.yaml in the current directory)")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug logging")
	cobra.OnInitialize(initServices)

	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
