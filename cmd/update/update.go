package update

import (
	"github.com/AaltoRSE/moat/internal/update"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateUpdateCmd creates the update command that updates the moat
// binary to the latest release.
func CreateUpdateCmd() *cobra.Command {

	// updateCmd represents the update command
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update moat to the latest release",
		Long: `Update moat to the latest release.

This command checks the upstream repository
(https://github.com/AaltoRSE/moat) for a newer release than the running
binary. When a newer release is available, it asks whether the binary
should be updated and whether the current binary should first be backed
up in the same directory. The release archive for the current operating
system and architecture is downloaded, its checksum is verified against
the release checksums, and the extracted binary replaces the running
binary.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := update.Update(); err != nil {
				log.Error().Msgf("could not update moat: %v", err)
			}
		},
	}

	return updateCmd
}
