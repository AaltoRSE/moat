package config

import (
	"github.com/spf13/cobra"
)

// CreateConfigCmd creates the config command and registers its subcommands.
func CreateConfigCmd() *cobra.Command {

	// ConfigCmd represents the config command
	var ConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Modify or view configuration settings",
		Long: `Handle configuration for moat.

Manage configuration settings.
You can modify, remove and view configuration settings.`,
	}

	ConfigCmd.AddCommand(CreateConfigSetCmd())
	ConfigCmd.AddCommand(CreateConfigAppendCmd())
	ConfigCmd.AddCommand(CreateConfigPrependCmd())
	ConfigCmd.AddCommand(CreateConfigShowCmd())
	ConfigCmd.AddCommand(CreateConfigEditCmd())
	return ConfigCmd
}
