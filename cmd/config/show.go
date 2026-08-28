package cmd

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/cobra"
)

// showCmd is the cobra command that prints the current configuration.
var showCmd = &cobra.Command{
	Use:     "show",
	Aliases: []string{"list", "view"},
	Short:   "Show the current moat configuration",
	Long: `Show the current moat configuration.

This command allows you to show the current moat configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(config.GetConfigAsString())
	},
}

// init registers the show command with the config command group.
func init() {
	configCmd.AddCommand(showCmd)
}
