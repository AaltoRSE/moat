package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-shellwords"
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

// SanitizeArgs normalizes a command line argument slice for execution.
//
// If args has a single element, it is parsed with shellwords.ParseWithEnvs so
// that quoted strings and VAR=value environment prefixes are handled
// correctly, and the resulting environment variables and command are
// returned. If args has more than one element, each argument is checked to
// contain no spaces; on success an empty environment and the original args
// are returned. An error is returned if any argument contains a space or if
// the single argument cannot be parsed.
func SanitizeArgs(args []string) (envVars []string, command []string, err error) {
	if len(args) == 1 {
		envVars, command, err = shellwords.ParseWithEnvs(args[0])
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse arguments with shellwords")
			return nil, nil, err
		}
		return envVars, command, nil
	}

	for _, arg := range args {
		if strings.Contains(arg, " ") {
			return nil, nil, fmt.Errorf("argument %q contains a space", arg)
		}
	}
	return []string{}, args, nil
}
