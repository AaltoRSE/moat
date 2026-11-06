package version

import (
	"testing"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/stretchr/testify/assert"
)

// TestCreateSemVersion verifies that CreateSemVersion parses semantic
// version strings with an optional leading "v" and an optional
// pre-release, build metadata, or git-describe suffix into their raw
// string, numeric version numbers, and commit hash.
func TestCreateSemVersion(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		want   types.SemVersion
		wantOK bool
	}{
		{"exact tag keeps the raw string", "v0.3.3", types.SemVersion{Raw: "v0.3.3", Version: [3]int{0, 3, 3}}, true},
		{"leading v is optional", "0.3.3", types.SemVersion{Raw: "0.3.3", Version: [3]int{0, 3, 3}}, true},
		{"surrounding whitespace is trimmed", " v0.3.3 ", types.SemVersion{Raw: "v0.3.3", Version: [3]int{0, 3, 3}}, true},
		{"git-describe suffix yields the commit hash", "v0.3.3-14-gb22ebb1", types.SemVersion{Raw: "v0.3.3-14-gb22ebb1", Version: [3]int{0, 3, 3}, Commit: "b22ebb1"}, true},
		{"build metadata after the commit is ignored", "v0.3.3-14-gb22ebb1+dirty", types.SemVersion{Raw: "v0.3.3-14-gb22ebb1+dirty", Version: [3]int{0, 3, 3}, Commit: "b22ebb1"}, true},
		{"pre-release suffix yields no commit", "v1.0.0-rc.1", types.SemVersion{Raw: "v1.0.0-rc.1", Version: [3]int{1, 0, 0}}, true},
		{"build metadata suffix yields no commit", "v1.0.0+build.5", types.SemVersion{Raw: "v1.0.0+build.5", Version: [3]int{1, 0, 0}}, true},
		{"empty version is not a semantic version", "", types.SemVersion{}, false},
		{"incomplete version is not a semantic version", "v0.3", types.SemVersion{}, false},
		{"extra version part is not a semantic version", "v0.3.3.1", types.SemVersion{}, false},
		{"non-numeric version is not a semantic version", "vX.Y.Z", types.SemVersion{}, false},
		{"non-numeric patch is not a semantic version", "v0.3.x-1-gdeadbee", types.SemVersion{}, false},
		{"garbage is not a semantic version", "garbage", types.SemVersion{}, false},
		{"negative version number is not a semantic version", "v0.3.-1", types.SemVersion{}, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := CreateSemVersion(c.raw)

			assert.Equal(t, c.wantOK, ok)
			assert.Equal(t, c.want, got)
		})
	}
}

// TestCompareSemVersion verifies that CompareSemVersion orders semantic
// versions by their numeric version numbers.
func TestCompareSemVersion(t *testing.T) {
	cases := []struct {
		name string
		a    types.SemVersion
		b    types.SemVersion
		want int
	}{
		{"equal versions compare as zero", types.SemVersion{Version: [3]int{0, 3, 3}}, types.SemVersion{Version: [3]int{0, 3, 3}}, 0},
		{"a greater patch version wins", types.SemVersion{Version: [3]int{0, 3, 4}}, types.SemVersion{Version: [3]int{0, 3, 3}}, 1},
		{"a smaller minor version loses", types.SemVersion{Version: [3]int{0, 3, 3}}, types.SemVersion{Version: [3]int{0, 4, 0}}, -1},
		{"the major version outranks minor and patch", types.SemVersion{Version: [3]int{1, 0, 0}}, types.SemVersion{Version: [3]int{0, 99, 99}}, 1},
		{"the minor version outranks the patch", types.SemVersion{Version: [3]int{0, 4, 0}}, types.SemVersion{Version: [3]int{0, 3, 99}}, 1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, CompareSemVersion(c.a, c.b))
		})
	}
}
