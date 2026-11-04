package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DirOpsTestSuite tests CreateDir: creation of missing directories, no-op
// on existing ones, and error reporting for invalid paths.
type DirOpsTestSuite struct {
	suite.Suite
	BaseDir string
}

// SetupTest creates a fresh empty base directory for each test.
func (suite *DirOpsTestSuite) SetupTest() {
	suite.BaseDir = suite.T().TempDir()
}

// TestCreateDirCreatesMissing tests that a missing directory, including any
// missing parent directories, is created.
func (suite *DirOpsTestSuite) TestCreateDirCreatesMissing() {
	dir := filepath.Join(suite.BaseDir, "parent", "child")

	require.NoError(suite.T(), CreateDir(dir))
	assert.DirExists(suite.T(), dir)
}

// TestCreateDirExistingIsNoOp tests that an existing directory is left
// untouched and no error is reported.
func (suite *DirOpsTestSuite) TestCreateDirExistingIsNoOp() {
	dir := filepath.Join(suite.BaseDir, "existing")
	require.NoError(suite.T(), os.Mkdir(dir, 0755))

	require.NoError(suite.T(), CreateDir(dir))
	assert.DirExists(suite.T(), dir)
}

// TestCreateDirError tests that an error is returned when the directory
// cannot be created because a parent path component is a regular file.
func (suite *DirOpsTestSuite) TestCreateDirError() {
	file := filepath.Join(suite.BaseDir, "file")
	require.NoError(suite.T(), os.WriteFile(file, []byte("content"), 0644))

	err := CreateDir(filepath.Join(file, "child"))
	assert.Error(suite.T(), err)
}

func TestDirOpsTestSuite(t *testing.T) {
	suite.Run(t, new(DirOpsTestSuite))
}
