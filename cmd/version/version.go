package version

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/version"
	"github.com/spf13/cobra"
)

// CreateVersionCmd creates the version command that prints the moat
// version.
func CreateVersionCmd() *cobra.Command {

	// versionCmd represents the version command
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the moat version",
		Long:  `Print the version of moat.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.MoatVersion)
		},
	}

	return versionCmd
}
