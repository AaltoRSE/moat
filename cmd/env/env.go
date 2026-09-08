package env

import (
	"github.com/spf13/cobra"
)

func CreateEnvCmd() *cobra.Command {

	// EnvCmd represents the env command
	var EnvCmd = &cobra.Command{
		Use:   "env",
		Short: "Manage moat environments",
		Long: `Manage moat environments.

This command allows you to manage your moat environments.
You can create and list environments.`,
	}

	EnvCmd.AddCommand(CreateEnvCreateCmd())
	EnvCmd.AddCommand(CreateEnvCopyCmd())
	EnvCmd.AddCommand(CreateEnvListCmd())
	EnvCmd.AddCommand(CreateEnvRemoveCmd())
	EnvCmd.AddCommand(CreateEnvShowCmd())
	return EnvCmd
}
