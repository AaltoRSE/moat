/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package root

import (
	"github.com/rs/zerolog/log"

	cmd_config "github.com/AaltoRSE/moat/cmd/config"
	"github.com/AaltoRSE/moat/cmd/env"
	"github.com/AaltoRSE/moat/cmd/run"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/logging"
	"github.com/spf13/cobra"
)

func CreateRootCmd() *cobra.Command {

	// RootCmd represents the base command when called without any subcommands
	rootCmd := &cobra.Command{
		Use:   "moat",
		Short: "A brief description of your application",
		Long: `Moat is an application for running AI agents in
	containerized environtment.`,
		DisableFlagParsing: false,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Init logging
			// Default level for this example is info, unless debug flag is present
			debug, _ := cmd.Flags().GetBool("debug")
			logging.InitLogging(debug)

			// Init configuration
			configFile, _ := cmd.Flags().GetString("config")
			log.Debug().Interface("configFile", configFile).Msg("Using config file: ")
			log.Debug().Interface("CmdConfig", config.CmdConfig).Msg("Using CmdConfig: ")
			cfg, err := config.InitConfig(configFile)
			if err != nil {
				log.Error().Err(err).Msgf("could not initialize config: %v", err)
			}
			config.CmdConfig = cfg
			return err
		},
	}
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file (default is $HOME/.config/moat/config.yaml or config.yaml in the current directory)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "enable debug logging")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(run.CreateRunCmd())
	rootCmd.AddCommand(env.CreateEnvCmd())
	rootCmd.AddCommand(cmd_config.CreateConfigCmd())

	return rootCmd
}
