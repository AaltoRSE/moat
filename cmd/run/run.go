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
			log.Error().Msgf("no arguments provided")
			return
		}
		// Get the first argument as the environment name
		envName := args[0]

		// If there is only one argument besides the environment name, parse it with shellwords to handle quoted strings correctly
		if len(args) == 2 {
			envVars, parsedArgs, err = shellwords.ParseWithEnvs(args[1])
			if err != nil {
				log.Error().Msgf("failed to parse arguments: %v", err)
				return
			}
		} else {
			envVars = []string{}
			parsedArgs = args[1:]
		}

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
		_, execErr := runtime.Run(env, parsedArgs, envVars)
		if execErr != nil {
			log.Error().Msgf("error executing command: %v", execErr)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(runCmd)
}
