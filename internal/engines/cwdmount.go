package engines

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

type cwdMount struct {
	source  string
	workdir string
}

// resolveCwdMount resolves the caller's directory once. Only entries using
// @cwd affect the selection of an explicit container working directory.
func resolveCwdMount(env types.MoatEnv, mountCWD bool) (cwdMount, error) {
	path, err := os.Getwd()
	if err != nil {
		return cwdMount{}, err
	}
	result := cwdMount{source: path}
	type mountEntry struct {
		dest    string
		cwd     bool
		overlay bool
	}
	var entries []mountEntry
	hasCWD := false
	for i, mounts := range [][]string{env.Mounts, env.ReadOnlyMounts, env.OverlayMounts} {
		for _, mount := range mounts {
			source, dest, explicit := strings.Cut(mount, ":")
			if source == utils.CwdMountSource && !hasCWD {
				hasCWD = true
				result.source, err = filepath.EvalSymlinks(path)
				if err != nil {
					return cwdMount{}, err
				}
			}
			if !explicit {
				dest = source
				if source == utils.CwdMountSource {
					dest = result.source
				}
			} else if i < 2 {
				dest, _, _ = strings.Cut(dest, ":")
			}
			entries = append(entries, mountEntry{dest, source == utils.CwdMountSource, i == 2})
		}
	}
	for i, entry := range entries {
		if !entry.cwd {
			continue
		}
		if !filepath.IsAbs(entry.dest) || entry.dest == "/" || strings.ContainsAny(entry.dest+result.source, "\r\n\x00") {
			return cwdMount{}, fmt.Errorf("invalid @cwd mount destination %q", entry.dest)
		}
		if entry.overlay && (strings.ContainsAny(entry.dest, ":,\\") || strings.ContainsFunc(entry.dest, unicode.IsSpace)) {
			return cwdMount{}, fmt.Errorf("overlay mount destination %q contains a mount delimiter or whitespace", entry.dest)
		}
		if entry.dest == os.Getenv("HOME") {
			return cwdMount{}, fmt.Errorf("@cwd mount destination %q conflicts with the fake home", entry.dest)
		}
		for j, other := range entries {
			if i == j {
				continue
			}
			if entry.dest == other.dest || (entry.overlay || other.overlay) &&
				(strings.HasPrefix(entry.dest, other.dest+"/") || strings.HasPrefix(other.dest, entry.dest+"/")) {
				return cwdMount{}, fmt.Errorf("@cwd mount destination %q conflicts with mount destination %q", entry.dest, other.dest)
			}
		}
		if mountCWD {
			if result.workdir != "" {
				return cwdMount{}, fmt.Errorf("mountcwd requires a single @cwd mapping, found %q and %q", result.workdir, entry.dest)
			}
			result.workdir = entry.dest
		}
	}
	return result, nil
}

func cwdMountDestination(mount, cwd string) (string, bool) {
	source, dest, explicit := strings.Cut(mount, ":")
	if source != utils.CwdMountSource {
		return "", false
	}
	if !explicit {
		dest = cwd
	}
	return dest, true
}
