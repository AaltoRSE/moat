/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/runtimes"
	"github.com/mattn/go-shellwords"
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
	DisableFlagParsing: false,

	Run: func(cmd *cobra.Command, args []string) {

		var (
			runtime    runtimes.Runtime
			err        error
			envVars    []string
			parsedArgs []string
		)

		log.Debug().Msgf("args: %v", args)

		// Return an error if no arguments are provided
		if len(args) == 0 {
			err := cmd.Help()
			if err != nil {
				log.Error().Msgf("Failed to display help: %v", err)
			}
			log.Error().Msgf("No arguments provided")
			return
		}
		// Get the first argument as the environment name
		envName := args[0]

		// If there is only one argument besides the environment name, parse it with shellwords to handle quoted strings correctly
		if len(args) == 2 {
			envVars, parsedArgs, err = shellwords.ParseWithEnvs(args[1])
			if err != nil {
				log.Error().Msgf("Failed to parse arguments: %v", err)
				return
			}
		} else {
			envVars = []string{}
			parsedArgs = args[1:]
		}

		// Get the environment with Config.GetEnv
		env, err := config.GetEnv(envName, true)
		if err != nil {
			log.Error().Msgf("Error with the environment %q: %v", envName, err)
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
		log.Debug().Msgf("Using runtime: %s", runtimeName)

		runtime, err = runtimes.GetRuntime(runtimeName)
		if err != nil {
			log.Error().Msgf("Failed to get a runtime: %v", err)
			return
		}

		log.Debug().Msgf("Executing runtime")
		_, execErr := runtime.Run(env, parsedArgs, envVars)
		if execErr != nil {
			log.Error().Msgf("Error executing command: %v", execErr)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(runCmd)
}
