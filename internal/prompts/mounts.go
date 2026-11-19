// Package prompts provides reusable, pre-defined interactive prompts that
// are asked from the user by user-facing moat commands.
package prompts

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// CheckAndCreateMounts checks that the paths referenced by the environment
// exist: the fake home directory and the sources of regular and read-only
// bind mounts. Mount sources may be directories or files:
// existing files are left untouched, while missing source paths are created
// as directories. Missing paths are created automatically when autoCreate
// is true; otherwise the user is asked whether they should be created. It
// returns an error when the user declines to create the missing paths, when
// their creation fails, or when a path is invalid (e.g. an empty mount
// source path).
func CheckAndCreateMounts(env types.MoatEnv, autoCreate bool) error {

	// Check that the fake home directory exists
	absFakeHome, err := utils.SanitizePath(env.Home)
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

	mounts := append(slices.Clone(env.Mounts), env.ReadOnlyMounts...)
	mounts = slices.DeleteFunc(mounts, func(mount string) bool {
		source, _, _ := strings.Cut(mount, ":")
		return source == utils.CwdMountSource
	})
	if utils.CheckMounts(mounts) {
		return nil
	}

	// Collect the missing source paths
	var missing []string
	seen := make(map[string]bool)
	for _, mount := range mounts {
		source := strings.Split(mount, ":")[0]
		if source == "" || seen[source] || utils.CheckPathExists(source) {
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

// CheckAndCreateOverlaySources checks the explicit host sources in mounts.
// Missing sources are created after confirmation unless autoCreate is true.
// The btrfs driver requires each source to be a subvolume.
func CheckAndCreateOverlaySources(mounts []string, driver string, autoCreate bool) error {
	for _, mount := range mounts {
		source, _, explicit := strings.Cut(mount, ":")
		if !explicit || source == utils.CwdMountSource {
			continue
		}
		info, err := os.Stat(source)
		if err == nil && !info.IsDir() {
			return fmt.Errorf("overlay source %q must be a directory", source)
		}
		if errors.Is(err, os.ErrNotExist) {
			if !autoCreate {
				ready, err := confirm(fmt.Sprintf("Create overlay source %s?", source))
				if err != nil {
					return err
				}
				if !ready {
					return fmt.Errorf("overlay source %q was not created", source)
				}
			}
		} else if err != nil {
			return err
		}
		if driver == "btrfs" {
			err = utils.EnsureBtrfsSubvolume(source)
		} else if info == nil {
			err = os.MkdirAll(source, 0o700)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
