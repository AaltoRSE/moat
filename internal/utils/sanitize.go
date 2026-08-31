package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func SanitizeFolderPath(path string) (string, error) {
	log.Debug().Msgf("Current environment: %s", os.Environ())
	log.Debug().Msgf("Sanitizing folder path: %s", path)
	absPath, err := filepath.Abs(os.ExpandEnv(path))
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Sanitized folder path: %s", absPath)
	return absPath, nil
}

func SanitizeMountsPaths(mounts []string) ([]string, error) {
	sanitizedMounts := make([]string, len(mounts))
	for i, mount := range mounts {
		// Split the mount string by colon to handle multiple paths in a single mount string
		paths := strings.Split(mount, ":")
		// Sanitize first entry of paths (the source path)
		sanitizedSource, err := SanitizeFolderPath(paths[0])
		if err != nil {
			return nil, err
		}
		// Reconstruct the mount string with sanitized source path
		if len(paths) > 1 {
			sanitizedMounts[i] = sanitizedSource + ":" + strings.Join(paths[1:], ":")
		} else {
			sanitizedMounts[i] = sanitizedSource
		}
	}
	return sanitizedMounts, nil
}
