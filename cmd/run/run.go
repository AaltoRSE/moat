/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	cmd_package "github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/runtimes"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
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

		// Get the environment name from the required flag
		envName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Error().Msgf("could not get environment name: %v", err)
			return
		}

		// Get the environment with Config.GetEnv
		env, err := config.GetEnv(cmd_package.CmdConfig, envName, true)
		if err != nil {
			log.Error().Msgf("Error with the environment %q: %v", envName, err)
			return
		}

		// Use the provided arguments as the command, falling back to the
		// environment's configured command when no arguments are given.
		command := args
		if len(command) == 0 {
			command = []string{env.Command}
		}

		// Return an error if no command is available to run
		if len(command) == 0 {
			err := cmd.Help()
			if err != nil {
				log.Error().Msgf("Failed to display help: %v", err)
			}
			log.Error().Msgf("No command provided")
			return
		}

		// Sanitize the command, parsing a single argument with shellwords to
		// handle quoted strings and environment prefixes correctly.
		envVars, parsedArgs, err = utils.SanitizeArgs(command)
		if err != nil {
			log.Error().Msgf("Failed to sanitize command: %v", err)
			return
		}

		// Check if the environment has a runtime specified

		var runtimeName string
		if env.Runtime != nil {
			runtimeName = *env.Runtime
		} else {
			runtimeName = cmd_package.CmdConfig.GetString("defaults.runtime")
		}

		// Log the runtime name
		log.Debug().Msgf("Using runtime: %s", runtimeName)

		runtime, err = runtimes.GetRuntime(cmd_package.CmdConfig, runtimeName)
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
	runCmd.Flags().StringP("name", "n", "", "Name of the environment to run the command in")

	if err := runCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	cmd_package.RootCmd.AddCommand(runCmd)
}
