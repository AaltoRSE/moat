// Package version holds build-time metadata for moat.
package version

// MoatVersion is the version of moat. It is set at build time via the
// linker flag -X github.com/AaltoRSE/moat/internal/version.MoatVersion
// and is empty when the binary is built without the flag.
var MoatVersion string
