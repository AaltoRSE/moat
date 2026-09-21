package runtime

import (
	"github.com/spf13/cobra"
)

// CreateRuntimeCmd creates the runtime command group and attaches all
// subcommands.
func CreateRuntimeCmd() *cobra.Command {

	// runtimeCmd represents the runtime command
	var runtimeCmd = &cobra.Command{
		Use:   "runtime",
		Short: "Manage moat runtimes",
		Long: `Manage moat runtimes.

This command allows you to manage your moat runtimes.
You can create and list runtimes.`,
	}

	runtimeCmd.AddCommand(CreateRuntimeCreateCmd())
	runtimeCmd.AddCommand(CreateRuntimeListCmd())
	return runtimeCmd
}
