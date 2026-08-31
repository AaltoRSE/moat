package cmd

import (
	"github.com/AaltoRSE/moat/cmd"
	"github.com/spf13/cobra"
)

// configCmd is the parent command for all configuration subcommands.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Modify or view configuration settings",
	Long: `Handle configuration for moat.

Manage configuration settings.
You can modify, remove and view configuration settings.`,
}

// init registers the config command with the root command.
func init() {
	cmd.RootCmd.AddCommand(configCmd)
}
