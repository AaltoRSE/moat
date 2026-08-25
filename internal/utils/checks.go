package utils

import (
	"os"
	"regexp"
	"strings"
)

// CheckFolderExists checks if a folder exists at the given path.
func CheckFolderExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// CheckEnvironmentName checks if the environment name is valid (non-empty and containing only alphanumeric characters and underscores).
func CheckEnvironmentName(name string) bool {
	var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validNamePattern.MatchString(name)
}

// CheckMounts checks if the provided mounts are valid (non-empty and existing directories).
func CheckMounts(mounts []string) bool {
	for _, mount := range mounts {
		// Split the mount string by colon to handle multiple paths in a single mount string
		paths := strings.Split(mount, ":")
		// Check if the first entry of paths (the source path) is valid
		if len(paths) == 0 || paths[0] == "" || !CheckFolderExists(paths[0]) {
			return false
		}
	}
	return true
}
