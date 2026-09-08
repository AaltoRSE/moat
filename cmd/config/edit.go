package config

import (
	"os"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateConfigEditCmd creates the edit subcommand for ConfigCmd.
func CreateConfigEditCmd() *cobra.Command {

	// editCmd is the cobra command that opens the currently active moat
	// configuration file in the editor defined by the $EDITOR environment
	// variable.
	var editCmd = &cobra.Command{
		Use:   "edit",
		Short: "Edit the moat configuration file in your $EDITOR",
		Long: `Edit the moat configuration file.

This command opens the currently active moat configuration file in the
editor defined by the $EDITOR environment variable. If no configuration
file exists yet, the global moat-config.yaml is opened.`,
		Run: func(cmd *cobra.Command, args []string) {
			editor := os.Getenv("EDITOR")
			if editor == "" {
				log.Fatal().Msgf("No editor specified in the environment. Set the EDITOR environment variable to your preferred editor (e.g. export EDITOR=vim)")
			}

			configFile := config.GetConfigFile(config.CmdConfig)
			if configFile == "" {
				log.Fatal().Msg("Could not determine the path of the configuration file")
			}

			log.Debug().Msgf("Opening configuration file %s in editor %s", configFile, editor)
			if err := utils.Run(utils.RunArgs{
				Command: editor,
				Args:    []string{configFile},
				PassEnv: true,
			}); err != nil {
				log.Fatal().Err(err).Msgf("Failed to run editor %q on %q", editor, configFile)
			}
		},
	}

	return editCmd
}
