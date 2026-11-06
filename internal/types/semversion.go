package types

// SemVersion holds a parsed semantic version of a moat release: the raw
// version string it was created from, the numeric major, minor, and
// patch version numbers, and the commit hash of a git-describe version
// suffix.
type SemVersion struct {
	// Raw is the string representation the version was created from,
	// for example v0.3.3-14-gb22ebb1.
	Raw string
	// Version holds the major, minor, and patch version numbers, in
	// this order.
	Version [3]int
	// Commit is the commit hash of a git-describe version suffix, for
	// example b22ebb1 for v0.3.3-14-gb22ebb1. It is empty for exact
	// release tags and for versions whose suffix is not a
	// git-describe commit suffix.
	Commit string
}
