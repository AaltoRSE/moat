package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ChecksTestSuite tests CheckFolderExists, CheckFileExists, CheckPathExists,
// and CheckMounts: existing directories, missing paths, regular files, and
// paths that cannot be stat'ed.
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

// TestCheckFileExists tests that CheckFileExists reports true only for
// regular files, and false for directories and missing paths.
func (suite *ChecksTestSuite) TestCheckFileExists() {
	dir := filepath.Join(suite.BaseDir, "dir")
	require.NoError(suite.T(), os.Mkdir(dir, 0755))
	file := filepath.Join(suite.BaseDir, "file")
	require.NoError(suite.T(), os.WriteFile(file, []byte("content"), 0644))

	assert.True(suite.T(), CheckFileExists(file))
	assert.False(suite.T(), CheckFileExists(dir))
	assert.False(suite.T(), CheckFileExists(filepath.Join(suite.BaseDir, "missing")))
}

// TestCheckPathExists tests that CheckPathExists reports true for both
// directories and regular files, and false for missing paths and paths
// whose parent component is a regular file.
func (suite *ChecksTestSuite) TestCheckPathExists() {
	dir := filepath.Join(suite.BaseDir, "dir")
	require.NoError(suite.T(), os.Mkdir(dir, 0755))
	file := filepath.Join(suite.BaseDir, "file")
	require.NoError(suite.T(), os.WriteFile(file, []byte("content"), 0644))

	assert.True(suite.T(), CheckPathExists(dir))
	assert.True(suite.T(), CheckPathExists(file))
	assert.False(suite.T(), CheckPathExists(filepath.Join(suite.BaseDir, "missing")))
	assert.False(suite.T(), CheckPathExists(filepath.Join(file, "child")))
}

// TestCheckMounts tests that CheckMounts reports true when every mount has
// a non-empty source path that exists as a directory or as a file, and
// false for empty, missing, and file-parented source paths.
func (suite *ChecksTestSuite) TestCheckMounts() {
	dir := filepath.Join(suite.BaseDir, "dir")
	require.NoError(suite.T(), os.Mkdir(dir, 0755))
	file := filepath.Join(suite.BaseDir, "file")
	require.NoError(suite.T(), os.WriteFile(file, []byte("content"), 0644))
	missing := filepath.Join(suite.BaseDir, "missing")

	assert.True(suite.T(), CheckMounts([]string{dir, dir + ":/dest"}))
	assert.True(suite.T(), CheckMounts([]string{file, file + ":/dest"}))
	assert.True(suite.T(), CheckMounts(nil))
	assert.False(suite.T(), CheckMounts([]string{""}))
	assert.False(suite.T(), CheckMounts([]string{"/dest"}))
	assert.False(suite.T(), CheckMounts([]string{missing}))
	assert.False(suite.T(), CheckMounts([]string{dir, missing}))
	assert.False(suite.T(), CheckMounts([]string{filepath.Join(file, "child")}))
}

func TestChecksTestSuite(t *testing.T) {
	suite.Run(t, new(ChecksTestSuite))
}
