package version

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/version"
	"github.com/spf13/cobra"
)

// CreateVersionCmd creates the version command that prints the moat
// version. The -a/--all flag includes the operating system and
// architecture the binary was built for in the output.
func CreateVersionCmd() *cobra.Command {
	var all bool

	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the moat version",
		Long:  `Print the version of moat.`,
		Run: func(cmd *cobra.Command, args []string) {
			if all {
				fmt.Println(version.FullVersion())
				return
			}
			fmt.Println(version.MoatVersion)
		},
	}

	versionCmd.Flags().BoolVarP(&all, "all", "a", false,
		"include the operating system and architecture in the version output")

	return versionCmd
}
