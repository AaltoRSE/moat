// Package version holds build-time metadata for moat.
package version

import (
	"regexp"
	"runtime"
	"strings"
)

// MoatVersion is the version of moat. It is set at build time via the
// linker flag -X github.com/AaltoRSE/moat/internal/version.MoatVersion
// and is empty when the binary is built without the flag.
var MoatVersion string

// MoatOS is the operating system the moat binary was built for. It is set at
// build time via the linker flag -X github.com/AaltoRSE/moat/internal/version.MoatOS
// and is empty when the binary is built without the flag.
var MoatOS string

// MoatArch is the processor architecture the moat binary was built for. It is
// set at build time via the linker flag -X github.com/AaltoRSE/moat/internal/version.MoatArch
// and is empty when the binary is built without the flag.
var MoatArch string

// MoatSource is the source repository of moat, for example
// github.com/AaltoRSE/moat. The moat update command checks this
// repository for a newer release. It is a variable so that tests can
// point it at a test server.
var MoatSource = "github.com/AaltoRSE/moat"

// imageTagPattern matches semantic version tags of the form vX.Y.Z.
var imageTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// FullVersion returns the full version string of the moat binary: the
// version ([MoatVersion]) followed by the platform identifier
// ([PlatformName]), for example v0.3.3-Linux_x86_64. The platform
// identifier is derived from the operating system and architecture the
// binary was built for ([MoatOS] and [MoatArch]) when they have been
// set, and from the platform moat is running on otherwise. When the
// binary was built without a version, FullVersion returns only the
// platform identifier.
func FullVersion() string {
	if MoatVersion == "" {
		return PlatformName()
	}
	return MoatVersion + "-" + PlatformName()
}

// PlatformName returns the platform identifier of the platform moat runs
// on: the title-cased operating system name followed by the uname-style
// architecture name, joined by an underscore, for example Linux_x86_64
// for linux and amd64. The build-time values [MoatOS] and [MoatArch] are
// used when they have been set; the operating system and architecture of
// the system moat is running on ([runtime.GOOS] and [runtime.GOARCH]) are
// used for the values that are not set. The identifier matches the
// platform part of the release archive names (for example,
// moat_Linux_x86_64.tar.gz). Architectures that are not in the mapping
// are reported as is.
func PlatformName() string {
	// titleOSName returns the operating system name with its first
	// letter upper-cased, for example Linux for linux.
	titleOSName := func(name string) string {
		return strings.ToUpper(name[:1]) + name[1:]
	}
	// archNames maps the Go architecture names stored in [MoatArch] to
	// the uname-style architecture names used in the full version string
	// and the release archive names. The mapping mirrors the naming of
	// the release archives; architectures that are not listed are
	// reported as is.
	archNames := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
	}
	// assetArchName returns the uname-style architecture name used in
	// the release archive names for the given Go architecture name;
	// architectures that are not in [archNames] are returned as is.
	assetArchName := func(arch string) string {
		if mapped, ok := archNames[arch]; ok {
			return mapped
		}
		return arch
	}

	osName, archName := runtime.GOOS, runtime.GOARCH
	if MoatOS != "" {
		osName = MoatOS
	}
	if MoatArch != "" {
		archName = MoatArch
	}
	return titleOSName(osName) + "_" + assetArchName(archName)
}

// ImageTag returns the container image tag derived from [MoatVersion]:
// the leading tag of the version string (for example, v0.3.3 for the
// git describe output v0.3.3-3-gcad91c9) when it matches the semantic
// version syntax vX.Y.Z, and latest otherwise (for example, when
// [MoatVersion] is empty or is not a semantic version).
func ImageTag() string {
	tag := strings.Split(MoatVersion, "-")[0]
	if imageTagPattern.MatchString(tag) {
		return tag
	}
	return "latest"
}
