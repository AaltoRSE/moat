package prompts_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/internal/prompts"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MountsTestSuite tests CheckAndCreateMounts: the check of the fake home
// directory and existing mount sources (directories and files) and the
// creation of missing ones, both automatically and after asking the user.
type MountsTestSuite struct {
	suite.Suite
}

// newTestEnv returns an environment whose home directory exists and whose
// mounts and read-only mounts point at the given directories. The home
// directory is removed when the test finishes.
func (suite *MountsTestSuite) newTestEnv(mounts, roMounts []string) types.MoatEnv {
	home := filepath.Join(tests.MoatTestDir, "prompts_mounts_home")
	require.NoError(suite.T(), os.MkdirAll(home, 0755))
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })
	return types.MoatEnv{
		Home:           home,
		Mounts:         mounts,
		ReadOnlyMounts: roMounts,
	}
}

// newMissingMount returns a mount string whose source directory does not
// exist yet, together with its source path. The source directory is
// removed when the test finishes.
func (suite *MountsTestSuite) newMissingMount(name string) (string, string) {
	source := filepath.Join(tests.MoatTestDir, name)
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(source)) })
	return source + ":/projects/missing", source
}

// newMissingHomeEnv returns an environment whose fake home directory does
// not exist and which has no mounts. The home directory is removed when
// the test finishes.
func (suite *MountsTestSuite) newMissingHomeEnv() types.MoatEnv {
	home := filepath.Join(tests.MoatTestDir, "prompts_missing_home")
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })
	return types.MoatEnv{Home: home}
}

// TestHomeAutoCreate tests that a missing fake home directory is created
// without prompting when autoCreate is true.
func (suite *MountsTestSuite) TestHomeAutoCreate() {
	env := suite.newMissingHomeEnv()
	absHome, err := utils.SanitizePath(env.Home)
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))

	assert.DirExists(suite.T(), absHome)
}

// TestHomePromptAccept tests that a missing fake home directory is created
// when the user accepts the prompt.
func (suite *MountsTestSuite) TestHomePromptAccept() {
	env := suite.newMissingHomeEnv()
	absHome, err := utils.SanitizePath(env.Home)
	require.NoError(suite.T(), err)
	tests.AnswerPrompt(suite.T(), "y")

	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, false))

	assert.DirExists(suite.T(), absHome)
}

// TestHomePromptDecline tests that no fake home directory is created and an
// error is returned when the user declines the prompt.
func (suite *MountsTestSuite) TestHomePromptDecline() {
	env := suite.newMissingHomeEnv()
	absHome, err := utils.SanitizePath(env.Home)
	require.NoError(suite.T(), err)
	tests.AnswerPrompt(suite.T(), "n")

	err = prompts.CheckAndCreateMounts(env, false)

	require.Error(suite.T(), err)
	assert.NoDirExists(suite.T(), absHome)
}

// TestHomeAndMountsAutoCreate tests that a missing fake home directory and
// missing mount sources are all created automatically when autoCreate is
// true.
func (suite *MountsTestSuite) TestHomeAndMountsAutoCreate() {
	mount, mountSource := suite.newMissingMount("prompts_missing_home_mount")
	env := suite.newMissingHomeEnv()
	env.Mounts = []string{mount}
	absHome, err := utils.SanitizePath(env.Home)
	require.NoError(suite.T(), err)

	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))

	assert.DirExists(suite.T(), absHome)
	assert.DirExists(suite.T(), mountSource)
}

// TestAllExist tests that no prompt is asked and no error is returned when
// all mount sources exist, regardless of autoCreate.
func (suite *MountsTestSuite) TestAllExist() {
	mount1 := filepath.Join(tests.MoatTestDir, "prompts_mounts_mount1")
	mount2 := filepath.Join(tests.MoatTestDir, "prompts_mounts_mount2")
	require.NoError(suite.T(), os.MkdirAll(mount1, 0755))
	require.NoError(suite.T(), os.MkdirAll(mount2, 0755))
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(mount1))
		require.NoError(suite.T(), utils.CleanupDir(mount2))
	})

	existing := suite.newTestEnv([]string{mount1}, []string{mount2})
	env := types.MoatEnv{
		Home:           existing.Home,
		Mounts:         []string{mount1},
		ReadOnlyMounts: []string{mount2},
	}
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, false))

	assert.DirExists(suite.T(), mount1)
	assert.DirExists(suite.T(), mount2)
}

// TestAutoCreate tests that missing mount and read-only mount sources are
// created without prompting when autoCreate is true.
func (suite *MountsTestSuite) TestAutoCreate() {
	mount, mountSource := suite.newMissingMount("prompts_missing_mount1")
	roMount, roMountSource := suite.newMissingMount("prompts_missing_mount2")

	env := suite.newTestEnv([]string{mount}, []string{roMount})
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))

	assert.DirExists(suite.T(), mountSource)
	assert.DirExists(suite.T(), roMountSource)
}

// TestPromptAccept tests that missing mount sources are created when the
// user accepts the prompt.
func (suite *MountsTestSuite) TestPromptAccept() {
	mount, mountSource := suite.newMissingMount("prompts_missing_mount")

	env := suite.newTestEnv([]string{mount}, nil)
	tests.AnswerPrompt(suite.T(), "y")
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, false))

	assert.DirExists(suite.T(), mountSource)
}

// TestPromptDecline tests that no directory is created and an error is
// returned when the user declines the prompt.
func (suite *MountsTestSuite) TestPromptDecline() {
	mount, mountSource := suite.newMissingMount("prompts_missing_mount")

	env := suite.newTestEnv([]string{mount}, nil)
	tests.AnswerPrompt(suite.T(), "n")
	err := prompts.CheckAndCreateMounts(env, false)

	require.Error(suite.T(), err)
	assert.NoDirExists(suite.T(), mountSource)
}

// TestInvalidEmptySource tests that a mount entry with an empty source path
// is reported as an error instead of being prompted for.
func (suite *MountsTestSuite) TestInvalidEmptySource() {
	env := suite.newTestEnv([]string{":/projects/invalid"}, nil)

	err := prompts.CheckAndCreateMounts(env, true)
	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid")

	tests.AnswerPrompt(suite.T(), "y")
	err = prompts.CheckAndCreateMounts(env, false)
	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "invalid")
}

// TestFileSource tests that a mount source that exists as a regular file
// is accepted and is left untouched, regardless of autoCreate.
func (suite *MountsTestSuite) TestFileSource() {
	filePath := filepath.Join(tests.MoatTestDir, "prompts_file_source_file")
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupFiles([]string{filePath})) })
	require.NoError(suite.T(), os.WriteFile(filePath, []byte("content"), 0644))

	env := suite.newTestEnv([]string{filePath + ":/projects/file"}, nil)
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, false))

	assert.FileExists(suite.T(), filePath)
	assert.NoDirExists(suite.T(), filePath)
}

// TestMixedFileAndDirectorySources tests that an existing file source and an
// existing directory source are both accepted, while a missing source is
// created as a directory.
func (suite *MountsTestSuite) TestMixedFileAndDirectorySources() {
	filePath := filepath.Join(tests.MoatTestDir, "prompts_mixed_file")
	dirPath := filepath.Join(tests.MoatTestDir, "prompts_mixed_dir")
	missingPath := filepath.Join(tests.MoatTestDir, "prompts_mixed_missing")
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupFiles([]string{filePath}))
		require.NoError(suite.T(), utils.CleanupDir(dirPath))
		require.NoError(suite.T(), utils.CleanupDir(missingPath))
	})
	require.NoError(suite.T(), os.WriteFile(filePath, []byte("content"), 0644))
	require.NoError(suite.T(), os.Mkdir(dirPath, 0755))

	env := suite.newTestEnv(
		[]string{filePath + ":/projects/file", missingPath},
		[]string{dirPath},
	)
	require.NoError(suite.T(), prompts.CheckAndCreateMounts(env, true))

	assert.FileExists(suite.T(), filePath)
	assert.NoDirExists(suite.T(), filePath)
	assert.DirExists(suite.T(), dirPath)
	assert.DirExists(suite.T(), missingPath)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestMountsTestSuite(t *testing.T) {
	suite.Run(t, new(MountsTestSuite))
}
