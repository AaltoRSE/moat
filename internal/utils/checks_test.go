package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ChecksTestSuite tests CheckFolderExists: existing directories, missing
// paths, regular files, and paths that cannot be stat'ed.
type ChecksTestSuite struct {
	suite.Suite
	BaseDir string
}

// SetupTest creates a fresh empty base directory for each test.
func (suite *ChecksTestSuite) SetupTest() {
	suite.BaseDir = suite.T().TempDir()
}

// TestCheckFolderExists tests that CheckFolderExists reports true only for
// existing directories, and false — instead of panicking — for missing
// paths, regular files, and paths whose parent component is a regular file.
func (suite *ChecksTestSuite) TestCheckFolderExists() {
	dir := filepath.Join(suite.BaseDir, "dir")
	require.NoError(suite.T(), os.Mkdir(dir, 0755))
	file := filepath.Join(suite.BaseDir, "file")
	require.NoError(suite.T(), os.WriteFile(file, []byte("content"), 0644))

	assert.True(suite.T(), CheckFolderExists(dir))
	assert.False(suite.T(), CheckFolderExists(filepath.Join(suite.BaseDir, "missing")))
	assert.False(suite.T(), CheckFolderExists(file))
	assert.False(suite.T(), CheckFolderExists(filepath.Join(file, "child")))
}

func TestChecksTestSuite(t *testing.T) {
	suite.Run(t, new(ChecksTestSuite))
}
