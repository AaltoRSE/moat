package utils

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// CreateDir creates the directory at the given path, including any missing
// parent directories, when it does not exist yet. It is a no-op when the
// directory already exists. It returns an error when the directory cannot
// be created.
func CreateDir(path string) error {
	if CheckFolderExists(path) {
		return nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Error().Err(err).Str("path", path).Msg("Failed to create directory")
		return err
	}
	log.Debug().Str("path", path).Msg("Directory created")
	return nil
}

// CleanupFiles removes the files at the given paths one by one. It is
// meant for cleaning up the known products of an operation: only the
// given paths are removed, and no directory is searched for additional
// files to delete. Paths that do not exist are skipped, so that
// products that were never (fully) created can also be listed. A path
// that names a directory is removed only when the directory is empty.
// It returns an error when any of the paths cannot be removed; all
// paths are attempted before it returns.
func CleanupFiles(files []string) error {
	var errs []error
	for _, file := range files {
		if err := os.Remove(file); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Error().Err(err).Str("path", file).Msg("Failed to remove file")
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// CleanupDir removes the directory at the given path, which must be
// empty. It is meant for removing a temporary directory after its known
// contents have been removed with [CleanupFiles]: a directory that
// still contains entries is reported as an error instead of being
// removed, so that files unknown to the caller are never deleted. A
// missing directory is a no-op. It returns an error when the path
// exists but is not a directory, when the directory is not empty, or
// when the directory cannot be removed.
func CleanupDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		log.Error().Err(err).Str("path", dir).Msg("Failed to inspect directory")
		return err
	}
	if len(entries) > 0 {
		err := fmt.Errorf("the directory %s is not empty", dir)
		log.Error().Err(err).Str("path", dir).Msg("Failed to remove directory")
		return err
	}
	if err := os.Remove(dir); err != nil {
		log.Error().Err(err).Str("path", dir).Msg("Failed to remove directory")
		return err
	}
	log.Debug().Str("path", dir).Msg("Directory removed")
	return nil
}
