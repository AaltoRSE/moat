package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-shellwords"
	"github.com/rs/zerolog/log"
)

// SanitizeFolderPath expands a leading `~` or `~/` prefix to the current
// user's home directory, expands environment variables in the path, and
// resolves the result to an absolute path. A `~` anywhere else in the path
// (including `~user` prefixes) is not expanded and is kept as a literal
// path element. It returns an error if the home directory cannot be
// determined or the absolute path cannot be resolved.
func SanitizeFolderPath(path string) (string, error) {
	log.Debug().Msgf("Current environment: %s", os.Environ())
	log.Debug().Msgf("Sanitizing folder path: %s", path)
	expanded, err := expandTilde(path)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(os.ExpandEnv(expanded))
	if err != nil {
		return "", err
	}
	log.Debug().Msgf("Sanitized folder path: %s", absPath)
	return absPath, nil
}

// expandTilde replaces a leading `~` or `~/` in path with the current
// user's home directory. It returns the path unchanged when it does not
// start with `~` or `~/`.
func expandTilde(path string) (string, error) {
	switch {
	case path == "~":
		return os.UserHomeDir()
	case strings.HasPrefix(path, "~/"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return home + path[1:], nil
	default:
		return path, nil
	}
}

// SanitizeMountsPaths sanitizes mount strings of the form `src`,
// `src:dest`, or `src:dest1:dest2`. Every path (source and destinations)
// is sanitized with [SanitizeFolderPath]: a leading `~` is expanded to the
// home directory, environment variables are expanded, and the result is
// resolved to an absolute path. It returns an error if any path cannot be
// sanitized.
func SanitizeMountsPaths(mounts []string) ([]string, error) {
	sanitizedMounts := make([]string, len(mounts))
	for i, mount := range mounts {
		// Split the mount string by colon to handle multiple paths in a single mount string
		paths := strings.Split(mount, ":")
		// Sanitize every entry of paths (the source path and the destination paths)
		for j, path := range paths {
			sanitizedPath, err := SanitizeFolderPath(path)
			if err != nil {
				return nil, err
			}
			paths[j] = sanitizedPath
		}
		// Reconstruct the mount string from the sanitized paths
		sanitizedMounts[i] = strings.Join(paths, ":")
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
		envVars, command, err = shellwords.ParseWithEnvs(os.ExpandEnv(args[0]))
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse arguments with shellwords")
			return nil, nil, err
		}
		log.Debug().Msgf("Sanitized command line arguments: %v", args)
		return envVars, command, nil
	}

	for _, arg := range args {
		if strings.Contains(arg, " ") {
			return nil, nil, fmt.Errorf("argument %q contains a space", arg)
		}
	}

	return []string{}, args, nil
}
