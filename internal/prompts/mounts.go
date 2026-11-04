// Package prompts provides reusable, pre-defined interactive prompts that
// are asked from the user by user-facing moat commands.
package prompts

import (
	"fmt"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// CheckAndCreateMounts checks that the directories referenced by the
// environment exist: the fake home directory and the source directories of
// the mounts and read-only mounts. Missing directories are created
// automatically when autoCreate is true; otherwise the user is asked whether
// they should be created. It returns an error when the user declines to
// create the missing directories, when their creation fails, or when a path
// is invalid (e.g. an empty mount source path or a source that exists but is
// not a directory).
func CheckAndCreateMounts(env types.MoatEnv, autoCreate bool) error {

	// Check that the fake home directory exists
	absFakeHome, err := utils.SanitizeFolderPath(env.Home)
	if err != nil {
		return err
	}
	if !utils.CheckFolderExists(absFakeHome) {
		log.Info().Str("absFakeHome", absFakeHome).Msg("Fake home directory does not exist")
		create := autoCreate
		if !create {
			ready, err := confirm("Do you want to create the fake home directory?")
			if err != nil {
				log.Error().Err(err).Msg("Error during confirmation prompt")
				return err
			}
			create = ready
		}
		if !create {
			fmt.Println("Aborting due to missing fake home directory.")
			return fmt.Errorf("missing fake home directory was not created")
		}
		if err := utils.CreateDir(absFakeHome); err != nil {
			return err
		}
		fmt.Printf("Fake home directory created: %s\n", absFakeHome)
	}

	// Check the sources of both mounts and read-only mounts in one go
	mounts := append(slices.Clone(env.Mounts), env.ReadOnlyMounts...)
	if utils.CheckMounts(mounts) {
		return nil
	}

	// Collect the missing source directories
	var missing []string
	seen := make(map[string]bool)
	for _, mount := range mounts {
		source := strings.Split(mount, ":")[0]
		if source == "" || seen[source] || utils.CheckFolderExists(source) {
			continue
		}
		seen[source] = true
		missing = append(missing, source)
	}

	// Mount entries that are not merely missing (e.g. an empty source path)
	// cannot be fixed by creating directories
	if len(missing) == 0 {
		return fmt.Errorf("one or more mount paths are invalid")
	}

	create := autoCreate
	if !create {
		fmt.Println("The following mount directories do not exist:")
		for _, dir := range missing {
			fmt.Printf("  %s\n", dir)
		}
		ready, err := confirm("Do you want to create the missing mount directories?")
		if err != nil {
			log.Error().Err(err).Msg("Error during confirmation prompt")
			return err
		}
		create = ready
	}
	if !create {
		fmt.Println("Aborting due to missing mount directories.")
		return fmt.Errorf("missing mount directories were not created")
	}

	// Create the missing source directories
	for _, dir := range missing {
		if err := utils.CreateDir(dir); err != nil {
			return err
		}
	}

	// Re-validate in case some mount entries are invalid rather than merely
	// missing.
	if !utils.CheckMounts(mounts) {
		return fmt.Errorf("one or more mount paths are invalid")
	}
	return nil
}
