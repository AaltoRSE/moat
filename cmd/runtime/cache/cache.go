package cache

import (
	"github.com/spf13/cobra"
)

// CreateCacheCmd creates the cache command group and attaches all
// subcommands.
func CreateCacheCmd() *cobra.Command {

	// cacheCmd represents the cache command
	var cacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "Manage the image cache of moat runtimes",
		Long: `Manage the image cache of moat runtimes.

This command allows you to inspect the container images that the
runtimes have cached locally.`,
	}

	cacheCmd.AddCommand(CreateCacheListCmd())
	return cacheCmd
}
