package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
