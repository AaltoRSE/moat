package init

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// CreateInitCmd creates the init command that initializes the moat
// configuration.
func CreateInitCmd() *cobra.Command {

	// outputPath is the path where the configuration is written. When
	// empty, the global configuration path
	// ($HOME/.config/moat/moat-config.yaml) is used.
	var outputPath string

	// initCmd is the cobra command that initializes the moat
	// configuration.
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize the moat configuration",
		Long: `Initialize the moat configuration.

Switches the configuration path of the configuration produced by
InitConfig to the output path (default:
$HOME/.config/moat/moat-config.yaml) and writes the configuration to
that path. When no output path is given and a configuration file has
already been found (at the global configuration path, in the current
directory, or via the --config flag), the configuration is left as it
is and its path is reported. An error is returned when the output path
is the same as the path of the found configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			changed, err := config.SetConfigPath(config.CmdConfig, outputPath)
			if err != nil {
				log.Fatal().Err(err).Msgf("failed to initialize the configuration: %v", err)
			}
			if !changed {
				return
			}
			if err := config.WriteConfig(config.CmdConfig); err != nil {
				log.Fatal().Err(err).Msgf("failed to write the configuration: %v", err)
			}
		},
	}

	initCmd.Flags().StringVarP(&outputPath, "output", "o", "",
		"output path of the configuration file (default is $HOME/.config/moat/moat-config.yaml)")

	return initCmd
}
