package cache

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/engines"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// CreateCacheListCmd creates the list subcommand for cacheCmd.
func CreateCacheListCmd() *cobra.Command {

	// listCmd represents the list command
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List the cached images of a moat runtime",
		Long: `List the cached images of a moat runtime.

This command allows you to view the container images that a runtime
has cached locally. For the apptainer type these are the .sif files in
the runtime's cache directory; for the docker type they are the image
references that docker reports for the repository of the runtime's
image, i.e. all of its locally stored tags.
Use the -n/--name flag to select the runtime; when the flag is not
given, the default runtime is used.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			// Fall back to the default runtime when no name is given
			if name == "" {
				name = config.CmdConfig.GetString("defaults.runtime")
			}

			engine, err := engines.GetRuntime(config.CmdConfig, name)
			if err != nil {
				log.Error().Msgf("Error with the runtime %q: %v", name, err)
				return
			}

			images, err := engine.CacheList()
			if err != nil {
				log.Error().Msgf("Error listing the cached images of runtime %q: %v", name, err)
				return
			}

			if len(images) == 0 {
				fmt.Printf("No cached images for runtime '%s'.\n", name)
				return
			}
			for _, image := range images {
				fmt.Println(image)
			}
		},
	}

	listCmd.Flags().StringP("name", "n", "", "Name of the runtime (default: the default runtime)")

	return listCmd
}
