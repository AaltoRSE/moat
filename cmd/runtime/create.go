package runtime

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/runtimes"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

// CreateRuntimeCreateCmd creates the create subcommand for runtimeCmd.
func CreateRuntimeCreateCmd() *cobra.Command {

	// createCmd represents the create command
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new moat runtime",
		Long: `Create a new moat runtime.

This command allows you to create and configure a new runtime.
The new runtime is stored in the top-level runtimes section of the
configuration and written to the config file after validation.

Examples:
  moat runtime create -n myrt --type apptainer --imageurl ghcr.io/aaltorse/moat:latest --cachedir $HOME/.cache/moat/images
  moat runtime create -n myrt --type apptainer --imageurl ghcr.io/aaltorse/moat:latest --cachedir $HOME/.cache/moat/images --mountcwd`,
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}
			runtimeType, err := cmd.Flags().GetString("type")
			if err != nil {
				log.Error().Msgf("could not get type flag: %v", err)
				return
			}
			imageUrl, err := cmd.Flags().GetString("imageurl")
			if err != nil {
				log.Error().Msgf("could not get imageurl flag: %v", err)
				return
			}
			cacheDir, err := cmd.Flags().GetString("cachedir")
			if err != nil {
				log.Error().Msgf("could not get cachedir flag: %v", err)
				return
			}
			passEnv, err := cmd.Flags().GetBool("passenv")
			if err != nil {
				log.Error().Msgf("could not get passenv flag: %v", err)
				return
			}
			mountCWD, err := cmd.Flags().GetBool("mountcwd")
			if err != nil {
				log.Error().Msgf("could not get mountcwd flag: %v", err)
				return
			}

			spec := types.RuntimeSpec{
				Type:     runtimeType,
				ImageUrl: imageUrl,
				CacheDir: cacheDir,
				PassEnv:  passEnv,
				MountCWD: mountCWD,
			}

			if err := runtimes.CreateRuntime(config.CmdConfig, name, spec); err != nil {
				log.Error().Msgf("could not create runtime: %v", err)
			}
		},
	}

	createCmd.Flags().StringP("name", "n", "", "Name of the runtime")
	createCmd.Flags().String("type", "", "Type of the runtime (e.g. apptainer)")
	createCmd.Flags().String("imageurl", "", "URL of the container image (required for the apptainer type)")
	createCmd.Flags().String("cachedir", "", "Directory for caching container images (required for the apptainer type)")
	createCmd.Flags().Bool("passenv", true, "Pass the environment variables to the container")
	createCmd.Flags().Bool("mountcwd", false, "Bind-mount the caller's current working directory into the container")

	if err := createCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := createCmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

	return createCmd
}
