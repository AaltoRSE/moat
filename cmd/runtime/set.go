package runtime

import (
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/runtime"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

// CreateRuntimeSetCmd creates the set subcommand for runtimeCmd.
func CreateRuntimeSetCmd() *cobra.Command {

	// setCmd is the cobra command that sets variables of an existing
	// runtime.
	var setCmd = &cobra.Command{
		Use:   "set",
		Short: "Set a variable of an existing moat runtime",
		Long: `Set a variable of an existing moat runtime.

This command allows you to update the variables of an existing
runtime. Only the runtime given by the -n/--name flag is modified;
all other runtimes and the rest of the configuration are left
untouched. Only the flags that are given are changed; all other
variables of the runtime keep their current values.

When the given runtime is a default runtime (for example apptainer),
a user runtime with the same name is created from the default
runtime's settings with the given changes applied, so that it
overwrites the default runtime in all lookups.

Examples:
  moat runtime set -n myrt --imageurl ghcr.io/aaltorse/moat:latest
  moat runtime set -n apptainer --mountcwd
  moat runtime set -n myrt --cachedir $HOME/.cache/moat/images --passenv
  moat runtime set -n apptainer --overlayowner external --overlaydriver kernel-overlayfs`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Error().Msgf("could not get name flag: %v", err)
				return
			}

			// Collect the flags that were explicitly given, in a fixed
			// order, so that the output is deterministic.
			var updates []types.RuntimeUpdate
			if cmd.Flags().Changed("type") {
				value, err := cmd.Flags().GetString("type")
				if err != nil {
					log.Error().Msgf("could not get type flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "type", Value: []string{value}})
			}
			if cmd.Flags().Changed("executable") {
				value, err := cmd.Flags().GetString("executable")
				if err != nil {
					log.Error().Msgf("could not get executable flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "executable", Value: []string{value}})
			}
			if cmd.Flags().Changed("imageurl") {
				value, err := cmd.Flags().GetString("imageurl")
				if err != nil {
					log.Error().Msgf("could not get imageurl flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "imageurl", Value: []string{value}})
			}
			if cmd.Flags().Changed("cachedir") {
				value, err := cmd.Flags().GetString("cachedir")
				if err != nil {
					log.Error().Msgf("could not get cachedir flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "cachedir", Value: []string{value}})
			}
			if cmd.Flags().Changed("passenv") {
				value, err := cmd.Flags().GetBool("passenv")
				if err != nil {
					log.Error().Msgf("could not get passenv flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "passenv", Value: []string{strconv.FormatBool(value)}})
			}
			if cmd.Flags().Changed("mountcwd") {
				value, err := cmd.Flags().GetBool("mountcwd")
				if err != nil {
					log.Error().Msgf("could not get mountcwd flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "mountcwd", Value: []string{strconv.FormatBool(value)}})
			}
			if cmd.Flags().Changed("overlayowner") {
				value, err := cmd.Flags().GetString("overlayowner")
				if err != nil {
					log.Error().Msgf("could not get overlayowner flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "overlayowner", Value: []string{value}})
			}
			if cmd.Flags().Changed("overlaydriver") {
				value, err := cmd.Flags().GetString("overlaydriver")
				if err != nil {
					log.Error().Msgf("could not get overlaydriver flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "overlaydriver", Value: []string{value}})
			}
			if cmd.Flags().Changed("overlaydir") {
				value, err := cmd.Flags().GetString("overlaydir")
				if err != nil {
					log.Error().Msgf("could not get overlaydir flag: %v", err)
					return
				}
				updates = append(updates, types.RuntimeUpdate{Field: "overlaydir", Value: []string{value}})
			}

			if len(updates) == 0 {
				log.Error().Msgf("no variables to set: give at least one of --type, --executable, --imageurl, --cachedir, --passenv, --mountcwd, --overlayowner, --overlaydriver or --overlaydir")
				return
			}

			if err := runtime.SetRuntime(config.CmdConfig, name, updates); err != nil {
				log.Error().Msgf("could not set runtime: %v", err)
			}
		},
	}

	setCmd.Flags().StringP("name", "n", "", "Name of the runtime")
	setCmd.Flags().String("type", "", "Type of the runtime (e.g. apptainer)")
	setCmd.Flags().String("executable", "", "Executable that runs the container (default: apptainer for the apptainer type, docker for the docker type)")
	setCmd.Flags().String("imageurl", "", "URL of the container image (required for the apptainer type)")
	setCmd.Flags().String("cachedir", "", "Directory for caching container images (required for the apptainer type)")
	setCmd.Flags().Bool("passenv", true, "Pass the environment variables to the container")
	setCmd.Flags().Bool("mountcwd", false, "Bind-mount the caller's current working directory into the container")
	setCmd.Flags().String("overlayowner", "", "Who provides the overlay mounts of the environments: internal or external — apptainer only")
	setCmd.Flags().String("overlaydriver", "", "Overlay filesystem driver: fuse-overlayfs, kernel-overlayfs or btrfs — apptainer only")
	setCmd.Flags().String("overlaydir", "", "Managed directory for externally owned overlays — apptainer only")

	if err := setCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return setCmd
}
