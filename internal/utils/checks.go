package utils

import (
	"os"
	"regexp"
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
