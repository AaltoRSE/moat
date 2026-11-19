package engines

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/rs/zerolog/log"
)

func pathWithin(parent, path string) bool {
	rel, err := filepath.Rel(parent, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// prepareOverlayTree publishes a complete lower or writable copy by rename.
// The caller holds the storage lock. Failed preparations can be retried.
func prepareOverlayTree(storage, name string, prepare func(string) error) (string, error) {
	target := filepath.Join(storage, name)
	info, err := os.Lstat(target)
	if err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("overlay tree %q must be a directory", target)
		}
		return target, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	staging, err := os.MkdirTemp(storage, ".partial-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(staging) }()
	tree := filepath.Join(staging, "tree")
	if err := prepare(tree); err != nil {
		removeTemporaryTree(tree)
		return "", err
	}
	if err := os.Rename(tree, target); err != nil {
		removeTemporaryTree(tree)
		return "", err
	}
	return target, nil
}

func snapshotOverlayTree(source, target, mode string) error {
	switch mode {
	case "nativefs":
		return nativeOverlaySnapshot(source, target)
	case "copy", "reflink":
		return copyOverlayTree(source, target, mode)
	case "auto":
		// Native snapshots and reflinks require compatible filesystems.
		for _, method := range []string{"nativefs", "reflink"} {
			if err := snapshotOverlayTree(source, target, method); err == nil {
				log.Debug().Str("method", method).Msg("Prepared OverlayFS snapshot")
				return nil
			} else {
				log.Debug().Err(err).Str("method", method).Msg("OverlayFS snapshot method unavailable")
			}
			removeTemporaryTree(target)
		}
		log.Debug().Str("method", "copy").Msg("Falling back to a copied OverlayFS snapshot")
		return copyOverlayTree(source, target, "copy")
	default:
		return fmt.Errorf("unknown overlayfssnapshot mode %q", mode)
	}
}

func nativeOverlaySnapshot(source, target string) error {
	// The bind into Apptainer is read-only. A writable btrfs snapshot
	// permits unprivileged cleanup on filesystems without subvolume deletion.
	if err := runTreeCommand("btrfs", []string{"subvolume", "snapshot", source, target}, true, "take a native btrfs snapshot"); err != nil {
		return err
	}
	// btrfs snapshots replace nested subvolumes with inode-2 placeholders.
	// Fall back to a file-level method rather than silently omitting files.
	return filepath.WalkDir(target, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Sys().(*syscall.Stat_t).Ino == 2 {
			return fmt.Errorf("native snapshot of %q omits nested subvolume %q", source, strings.TrimPrefix(path, target+"/"))
		}
		return nil
	})
}

func overlayCopyArgs(mode string) []string {
	reflink := "never"
	if mode == "reflink" {
		reflink = "always"
	}
	return []string{"-a", "--no-preserve=ownership", "--reflink=" + reflink, "--"}
}

func copyOverlayTree(source, target, mode string) error {
	args := append(overlayCopyArgs(mode), source+"/.", target)
	return runTreeCommand("cp", args, true, mode+" overlay tree")
}

func (f *ApptainerRuntime) copyImageTree(imagePath, source, target, mode string, passEnv bool) error {
	if err := os.Mkdir(target, 0o700); err != nil {
		return err
	}
	dest := "/" + filepath.Base(filepath.Dir(target))
	args := []string{"exec", "--contain", "--no-home", "--no-mount", "cwd", "--no-eval",
		"--mount", mountSpec(target, dest), imagePath, "cp"}
	args = append(args, overlayCopyArgs(mode)...)
	args = append(args, source+"/.", dest)
	return runTreeCommand(f.Executable, args, passEnv, mode+" image directory")
}

func runTreeCommand(command string, args []string, passEnv bool, action string) error {
	output, err := utils.RunCapture(utils.RunArgs{Command: command, Args: args, PassEnv: passEnv})
	if err != nil {
		return fmt.Errorf("%s: %w: %s", action, err, strings.TrimSpace(output))
	}
	return nil
}

func removeTemporaryTree(path string) {
	if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
		return
	}
	// Removing a btrfs subvolume directly avoids walking its contents.
	_, _ = utils.RunCapture(utils.RunArgs{Command: "btrfs", Args: []string{"subvolume", "delete", path}, PassEnv: true})
	if err := os.RemoveAll(path); err != nil {
		// Archived modes can leave directories without owner write access.
		_ = filepath.WalkDir(path, func(dir string, entry os.DirEntry, err error) error {
			if err == nil && entry.IsDir() {
				_ = os.Chmod(dir, 0o700)
			}
			return nil
		})
		if err := os.RemoveAll(path); err != nil {
			log.Warn().Err(err).Str("path", path).Msg("Could not remove temporary overlay tree")
		}
	}
}
