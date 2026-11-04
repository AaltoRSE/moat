package utils

import (
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
