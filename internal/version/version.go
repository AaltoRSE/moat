// Package version holds build-time metadata for moat.
package version

import (
	"regexp"
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

// imageTagPattern matches semantic version tags of the form vX.Y.Z.
var imageTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// archNames maps the Go architecture names stored in [MoatArch] to the
// uname-style architecture names used in the full version string. The
// mapping mirrors the naming of the release archives; architectures that
// are not listed are reported as is.
var archNames = map[string]string{
	"amd64": "x86_64",
	"386":   "i386",
}

// FullVersion returns the full version string of the moat binary: the
// version ([MoatVersion]) followed by the operating system and architecture
// the binary was built for ([MoatOS] and [MoatArch]), for example
// v0.3.3-Linux_x86_64. The operating system is title-cased and the
// architecture is reported with its uname-style name (for example, amd64 as
// x86_64). Parts that are empty are omitted, so FullVersion returns
// [MoatVersion] as is when the binary was built without the operating
// system and architecture information.
func FullVersion() string {
	platform := ""
	if MoatOS != "" {
		platform = strings.ToUpper(MoatOS[:1]) + MoatOS[1:]
	}
	if MoatArch != "" {
		arch := MoatArch
		if mapped, ok := archNames[arch]; ok {
			arch = mapped
		}
		if platform != "" {
			platform += "_"
		}
		platform += arch
	}

	full := MoatVersion
	if platform != "" {
		if full != "" {
			full += "-"
		}
		full += platform
	}
	return full
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
