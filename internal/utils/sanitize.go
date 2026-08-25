package utils

import (
	"path/filepath"
	"strings"
)

func SanitizeFolderPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
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
