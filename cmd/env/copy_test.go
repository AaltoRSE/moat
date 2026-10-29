package env_test

import (
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopyWithMountCWD tests that env copy sets working-directory mounting in
// the copy when the --mountcwd flag is given with an empty value (empty means
// true), even though the source environment does not have it set.
func (suite *EnvTestSuite) TestCopyWithMountCWD() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv", "--mountcwd="})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), createdEnv.MountCWD)
	assert.True(suite.T(), *createdEnv.MountCWD)

	// The source environment must be untouched
	sourceEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Nil(suite.T(), sourceEnv.MountCWD)
}

// TestCopyWithMountCWDFalse tests that env copy disables working-directory
// mounting in the copy with --mountcwd=false, overriding the enabled setting
// of the source environment.
func (suite *EnvTestSuite) TestCopyWithMountCWDFalse() {
	// Enable mountcwd in the source environment
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--mountcwd="})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Copy the environment, explicitly disabling working-directory mounting
	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv", "--mountcwd=false"})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err = capture.StopCapture()
	require.NoError(suite.T(), err)

	// The copy must have mountcwd disabled while the source keeps it enabled
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), createdEnv.MountCWD)
	assert.False(suite.T(), *createdEnv.MountCWD)
	sourceEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), sourceEnv.MountCWD)
	assert.True(suite.T(), *sourceEnv.MountCWD)
}

// TestCopyInheritsOtherFields tests that env copy preserves the fields of the
// source environment that are not overridden by flags, including a mountcwd
// setting that is enabled in the source.
func (suite *EnvTestSuite) TestCopyInheritsOtherFields() {
	// Enable mountcwd in the source environment
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--mountcwd="})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Copy the environment without giving --mountcwd
	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv"})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err = capture.StopCapture()
	require.NoError(suite.T(), err)

	// The copy must have inherited the source's mountcwd setting
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), createdEnv.MountCWD)
	assert.True(suite.T(), *createdEnv.MountCWD)
	assert.Equal(suite.T(), suite.BaseEnv.Home, createdEnv.Home)
}

// TestCopyWithRoMounts tests that env copy overrides the source environment's
// read-only mounts when the --ro-mount flag is provided.
func (suite *EnvTestSuite) TestCopyWithRoMounts() {
	baseDir := filepath.Join(tests.MoatTestDir, "copy_ro_mounts")
	home := filepath.Join(baseDir, "home")
	roMount := filepath.Join(baseDir, "romount")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), os.RemoveAll(baseDir))
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv", "--home", home, "--ro-mount", roMount, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The home and read-only mount source directories must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), roMount)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{roMount}, createdEnv.ReadOnlyMounts)
}
