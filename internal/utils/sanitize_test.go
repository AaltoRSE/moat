package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SanitizeTestSuite tests SanitizePath and SanitizeMountsPaths:
// tilde expansion, environment variable expansion (including the UID
// fallback), and resolution of relative paths to absolute ones.
type SanitizeTestSuite struct {
	suite.Suite
	HomeDir string
	WorkDir string
}

// SetupTest isolates the home directory and the working directory so that
// tilde expansion and relative path resolution are deterministic.
func (suite *SanitizeTestSuite) SetupTest() {
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)

	suite.WorkDir = suite.T().TempDir()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), os.Chdir(suite.WorkDir))
	suite.T().Cleanup(func() { _ = os.Chdir(wd) })
}

func (suite *SanitizeTestSuite) TestTildeExpansion() {
	cases := []struct {
		path string
		want string
	}{
		{"~", suite.HomeDir},
		{"~/", suite.HomeDir},
		{"~/sub/dir", filepath.Join(suite.HomeDir, "sub/dir")},
		// A `~` that is not a leading home prefix is not expanded; the
		// result is a relative path resolved against the working
		// directory.
		{"~user/sub", filepath.Join(suite.WorkDir, "~user/sub")},
		{"~a", filepath.Join(suite.WorkDir, "~a")},
		{"sub/~/nested", filepath.Join(suite.WorkDir, "sub/~/nested")},
	}
	for _, tc := range cases {
		got, err := SanitizePath(tc.path)
		require.NoError(suite.T(), err, tc.path)
		assert.Equal(suite.T(), tc.want, got, tc.path)
	}
}

func (suite *SanitizeTestSuite) TestEnvVarExpansion() {
	varDir := suite.T().TempDir()
	suite.T().Setenv("MOAT_SANITIZER_DIR", varDir)

	got, err := SanitizePath("$MOAT_SANITIZER_DIR/sub")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(varDir, "sub"), got)

	// An unset variable expands to the empty string, so the leading
	// slash makes the result an absolute path.
	got, err = SanitizePath("$MOAT_UNSET_VAR/sub")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "/sub", got)
}

func (suite *SanitizeTestSuite) TestUIDExpansion() {
	// Remove any UID from the environment (restoring the original value
	// at cleanup) so that the fallback to the numeric user ID applies.
	if orig, ok := os.LookupEnv("UID"); ok {
		suite.T().Setenv("UID", orig)
		require.NoError(suite.T(), os.Unsetenv("UID"))
	}

	// $UID expands to the numeric user ID, yielding a relative path
	// resolved against the working directory.
	got, err := SanitizePath("$UID/data")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(),
		filepath.Join(suite.WorkDir, fmt.Sprintf("%d/data", os.Getuid())), got)

	// The braced form expands the same way.
	got, err = SanitizePath("${UID}/data")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(),
		filepath.Join(suite.WorkDir, fmt.Sprintf("%d/data", os.Getuid())), got)
}

func (suite *SanitizeTestSuite) TestUIDFromEnvironment() {
	// A UID set in the environment overrides the numeric user ID.
	suite.T().Setenv("UID", "4242")

	got, err := SanitizePath("$UID/data")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(suite.WorkDir, "4242/data"), got)
}

func (suite *SanitizeTestSuite) TestAbsoluteAndRelativePaths() {
	cases := []struct {
		path string
		want string
	}{
		{"/tmp", "/tmp"},
		{suite.WorkDir, suite.WorkDir},
		{"sub/dir", filepath.Join(suite.WorkDir, "sub/dir")},
		{".", suite.WorkDir},
		{"", suite.WorkDir},
	}
	for _, tc := range cases {
		got, err := SanitizePath(tc.path)
		require.NoError(suite.T(), err, tc.path)
		assert.Equal(suite.T(), tc.want, got, tc.path)
	}
}

func (suite *SanitizeTestSuite) TestSanitizeMountsPaths() {
	varDir := suite.T().TempDir()
	suite.T().Setenv("MOAT_SANITIZER_DIR", varDir)

	mounts := []string{
		"~/proj",
		"~/proj:/workspace",
		"$MOAT_SANITIZER_DIR/src:dest",
		"$MOAT_SANITIZER_DIR:$MOAT_SANITIZER_DIR",
		"src:$MOAT_SANITIZER_DIR:$MOAT_SANITIZER_DIR/nested",
		"src:~/dest",
		"plain",
	}
	want := []string{
		filepath.Join(suite.HomeDir, "proj"),
		filepath.Join(suite.HomeDir, "proj") + ":/workspace",
		filepath.Join(varDir, "src") + ":" + filepath.Join(suite.WorkDir, "dest"),
		varDir + ":" + varDir,
		filepath.Join(suite.WorkDir, "src") + ":" + varDir + ":" + varDir + "/nested",
		filepath.Join(suite.WorkDir, "src") + ":" + filepath.Join(suite.HomeDir, "dest"),
		filepath.Join(suite.WorkDir, "plain"),
	}

	got, err := SanitizeMountsPaths(mounts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), want, got)
}

func TestSanitizeTestSuite(t *testing.T) {
	suite.Run(t, new(SanitizeTestSuite))
}
