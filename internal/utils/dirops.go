package utils

import (
	"os"
	"strings"

	"github.com/rs/zerolog/log"
)

// CreateMountDirs creates the source directories of mounts that do
// not exist yet. Mount entries may use the "source:destination" form; only
// the source path is created. Mount entries without a source path are
// ignored and remain invalid.
func CreateMountDirs(mounts []string) error {
	for _, mount := range mounts {
		// Split the mount string by colon to handle multiple paths in a single mount string
		source := strings.Split(mount, ":")[0]
		if source == "" || CheckFolderExists(source) {
			continue
		}
		if err := os.MkdirAll(source, 0755); err != nil {
			log.Error().Err(err).Str("source", source).Msg("Failed to create mount directory")
			return err
		}
		log.Debug().Str("source", source).Msg("Mount directory created")
	}
	return nil
}
