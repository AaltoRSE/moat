/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/runtimes"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a command in moat",
	Long: `Run a command in moat.

This command allows you to run a specific command within the moat environment.`,
	DisableFlagParsing: true,

	Run: func(cmd *cobra.Command, args []string) {
		log.Print("run called with args:", args)

		var runtime runtimes.Runtime
		var err error

		// Print help if one of the arguments is --help or -h
		for _, arg := range args {
			if arg == "--help" || arg == "-h" {
				if err := cmd.Help(); err != nil {
					log.Error().Msgf("could not print help: %v", err)
				}
				return
			}
		}

		// Return an error if no arguments are provided
		if len(args) == 0 {
			log.Error().Msgf("no arguments provided")
			return
		}

		// Get the first argument as the environment name
		envName := args[0]

		// Get the environment with Config.GetEnv
		env, err := config.GetEnv(envName)
		if err != nil {
			log.Error().Msgf("could not get environment %q: %v", envName, err)
			return
		}

		// Check if the environment has a runtime specified

		var runtimeName string
		if env.Runtime != "" {
			runtimeName = env.Runtime
		} else {
			runtimeName = viper.GetString("defaults.runtime")
		}

		// Log the runtime name
		log.Print("Using runtime: ", runtimeName)

		runtime, err = runtimes.GetRuntime(runtimeName)
		if err != nil {
			log.Error().Msgf("failed to get a runtime: %v", err)
			return
		}

		log.Print("Executing runtime")
		_, execErr := runtime.Run(env, args[1:])
		if execErr != nil {
			log.Error().Msgf("error executing command: %v", execErr)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
