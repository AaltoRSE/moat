package version

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// runtimeOSName returns the title-cased operating system name reported
// by the Go runtime, mirroring the formatting of PlatformName.
func runtimeOSName() string {
	return strings.ToUpper(runtime.GOOS[:1]) + runtime.GOOS[1:]
}

// runtimeArchName returns the uname-style architecture name for the
// architecture reported by the Go runtime, mirroring the mapping of
// PlatformName.
func runtimeArchName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	}
	return runtime.GOARCH
}

// TestPlatformName verifies that PlatformName builds the platform
// identifier from the build-time operating system and architecture when
// they have been set, and falls back to the values reported by the Go
// runtime for the values that are not set.
func TestPlatformName(t *testing.T) {
	cases := []struct {
		name string
		os   string
		arch string
		want string
	}{
		{"build-time values are used when set", "linux", "amd64", "Linux_x86_64"},
		{"darwin arch name is kept", "darwin", "arm64", "Darwin_arm64"},
		{"386 is mapped to the uname-style name", "linux", "386", "Linux_i386"},
		{"unmapped arch names are used as is", "linux", "riscv64", "Linux_riscv64"},
		{"unset operating system falls back to the runtime", "", "amd64", runtimeOSName() + "_x86_64"},
		{"unset architecture falls back to the runtime", "darwin", "", "Darwin_" + runtimeArchName()},
		{"unset values fall back to the runtime", "", "", runtimeOSName() + "_" + runtimeArchName()},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			originalOS, originalArch := MoatOS, MoatArch
			t.Cleanup(func() { MoatOS, MoatArch = originalOS, originalArch })
			MoatOS, MoatArch = c.os, c.arch

			assert.Equal(t, c.want, PlatformName())
		})
	}
}

// TestFullVersion verifies that FullVersion combines MoatVersion with
// the platform identifier from PlatformName: the platform is formatted
// in the uname style from the build-time values when they are set, and
// falls back to the values reported by the Go runtime otherwise. The
// dash between the version and the platform is omitted when the version
// is empty.
func TestFullVersion(t *testing.T) {
	cases := []struct {
		name    string
		version string
		os      string
		arch    string
		want    string
	}{
		{"empty values yield the runtime platform", "", "", "", runtimeOSName() + "_" + runtimeArchName()},
		{"only the version is joined with the runtime platform", "v0.3.3", "", "", "v0.3.3-" + runtimeOSName() + "_" + runtimeArchName()},
		{"linux on amd64 uses the uname-style arch name", "v0.3.3", "linux", "amd64", "v0.3.3-Linux_x86_64"},
		{"darwin on arm64 keeps the arch name", "v0.3.3-3-gcad91c9", "darwin", "arm64", "v0.3.3-3-gcad91c9-Darwin_arm64"},
		{"386 uses the uname-style arch name", "v0.3.3", "linux", "386", "v0.3.3-Linux_i386"},
		{"unmapped arch names are used as is", "v0.3.3", "linux", "riscv64", "v0.3.3-Linux_riscv64"},
		{"arch without os falls back to the runtime operating system", "v0.3.3", "", "amd64", "v0.3.3-" + runtimeOSName() + "_x86_64"},
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
