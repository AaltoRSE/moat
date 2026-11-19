package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnsureBtrfsSubvolume creates a btrfs subvolume at dir and any missing parent
// directories. It returns an error if an existing path is not a subvolume or
// if btrfs cannot create or verify the subvolume.
func EnsureBtrfsSubvolume(dir string) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
			return err
		}
		output, createErr := RunCapture(RunArgs{Command: "btrfs", Args: []string{"subvolume", "create", dir}})
		if createErr == nil {
			return nil
		}
		// A simultaneous run may have created the directory in the meantime.
		if _, statErr := os.Stat(dir); statErr != nil {
			return fmt.Errorf("create btrfs subvolume %q: %w: %s", dir, createErr, strings.TrimSpace(output))
		}
	} else if err != nil {
		return err
	}

	// Querying the root's flags works without the privileges needed to
	// enumerate subvolumes and rejects ordinary directories.
	output, err := RunCapture(RunArgs{Command: "btrfs", Args: []string{"property", "get", "-t", "subvol", dir, "ro"}})
	if err != nil {
		return fmt.Errorf("check whether %q is a btrfs subvolume: %w: %s", dir, err, strings.TrimSpace(output))
	}
	if strings.TrimSpace(output) == "ro=true" {
		return fmt.Errorf("btrfs subvolume %q is read-only", dir)
	}
	return nil
}
