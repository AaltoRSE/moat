package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// CreateConfigShowDefaultsCmd creates the show-defaults subcommand for
// ConfigCmd.
func CreateConfigShowDefaultsCmd() *cobra.Command {

	// showDefaultsCmd is the cobra command that prints the default moat
	// configuration.
	var showDefaultsCmd = &cobra.Command{
		Use:   "show-defaults",
		Short: "Show the default moat configuration",
		Long: `Show the default moat configuration.

This command allows you to show the default moat configuration.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(config.GetConfigAsString(config.CreateDefaultConfig()))
		},
	}

	return showDefaultsCmd
}
