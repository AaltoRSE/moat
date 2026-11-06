package version

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/AaltoRSE/moat/internal/types"
)

// commitPattern matches the commit suffix of a git-describe version, for
// example 14-gb22ebb1 in v0.3.3-14-gb22ebb1, capturing the commit hash.
var commitPattern = regexp.MustCompile(`^\d+-g([0-9a-fA-F]+)$`)

// CreateSemVersion parses a semantic version string with an optional
// leading "v" and an optional pre-release, build metadata, or
// git-describe suffix (for example v0.3.3, 0.3.3, or
// v0.3.3-14-gb22ebb1) into a [types.SemVersion]. Raw holds the given
// version string, Version holds the numeric major, minor, and patch
// version numbers, and Commit holds the commit hash of a git-describe
// suffix (for example b22ebb1 for v0.3.3-14-gb22ebb1); any other
// suffix is ignored. It reports false when the string is not a semantic
// version.
func CreateSemVersion(raw string) (types.SemVersion, bool) {
	raw = strings.TrimSpace(raw)
	rest := strings.TrimPrefix(raw, "v")
	commit := ""
	if i := strings.IndexAny(rest, "-+"); i >= 0 {
		suffix := rest[i+1:]
		rest = rest[:i]
		// Build metadata is not part of the commit suffix.
		if j := strings.Index(suffix, "+"); j >= 0 {
			suffix = suffix[:j]
		}
		if m := commitPattern.FindStringSubmatch(suffix); m != nil {
			commit = m[1]
		}
	}
	parts := strings.Split(rest, ".")
	if len(parts) != 3 {
		return types.SemVersion{}, false
	}
	var numbers [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return types.SemVersion{}, false
		}
		numbers[i] = n
	}
	return types.SemVersion{Raw: raw, Version: numbers, Commit: commit}, true
}

// CompareSemVersion compares two semantic versions by their numeric
// version numbers and returns -1, 0, or 1 as a is less than, equal to,
// or greater than b.
func CompareSemVersion(a, b types.SemVersion) int {
	for i := range a.Version {
		switch {
		case a.Version[i] < b.Version[i]:
			return -1
		case a.Version[i] > b.Version[i]:
			return 1
		}
	}
	return 0
}
