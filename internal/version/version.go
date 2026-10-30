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

// imageTagPattern matches semantic version tags of the form vX.Y.Z.
var imageTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

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
