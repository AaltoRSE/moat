package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFullVersion verifies that FullVersion combines MoatVersion, MoatOS,
// and MoatArch into the full version string, formatting the operating
// system and architecture in the uname style and omitting empty parts.
func TestFullVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{"empty values yield an empty string", "", "", "", ""},
		{"only the version is used as is", "v0.3.3", "", "", "v0.3.3"},
		{"linux on amd64 uses the uname-style arch name", "v0.3.3", "linux", "amd64", "v0.3.3-Linux_x86_64"},
		{"darwin on arm64 keeps the arch name", "v0.3.3-3-gcad91c9", "darwin", "arm64", "v0.3.3-3-gcad91c9-Darwin_arm64"},
		{"386 uses the uname-style arch name", "v0.3.3", "linux", "386", "v0.3.3-Linux_i386"},
		{"unmapped arch names are used as is", "v0.3.3", "linux", "riscv64", "v0.3.3-Linux_riscv64"},
		{"arch without os is appended to the version", "v0.3.3", "", "amd64", "v0.3.3-x86_64"},
		{"os and arch without version omit the dash", "", "linux", "amd64", "Linux_x86_64"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			originalVersion, originalOS, originalArch := MoatVersion, MoatOS, MoatArch
			t.Cleanup(func() { MoatVersion, MoatOS, MoatArch = originalVersion, originalOS, originalArch })
			MoatVersion, MoatOS, MoatArch = c.version, c.os, c.arch

			assert.Equal(t, c.want, FullVersion())
		})
	}
}

// TestImageTag verifies that ImageTag returns the leading tag of
// MoatVersion when it is a semantic version tag, and falls back to
// latest otherwise.
func TestImageTag(t *testing.T) {
	cases := []struct {
		name    string
		version string
		want    string
	}{
		{"empty version falls back to latest", "", "latest"},
		{"exact tag is used as is", "v0.3.3", "v0.3.3"},
		{"leading tag of describe output is used", "v0.3.3-3-gcad91c9", "v0.3.3"},
		{"version without leading v falls back to latest", "0.3.3", "latest"},
		{"incomplete version falls back to latest", "v0.3", "latest"},
		{"non-numeric version falls back to latest", "v0.3.x-1-gdeadbee", "latest"},
		{"non-semantic version falls back to latest", "release", "latest"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			original := MoatVersion
			t.Cleanup(func() { MoatVersion = original })
			MoatVersion = c.version

			assert.Equal(t, c.want, ImageTag())
		})
	}
}
