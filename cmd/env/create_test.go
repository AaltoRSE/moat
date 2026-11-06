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

// TestCreateHomeOnly tests that env create registers a new environment with
// only a home directory (no mounts) and creates the fake home directory.
func (suite *EnvTestSuite) TestCreateHomeOnly() {
	home := filepath.Join(tests.MoatTestDir, "create_home_only_home")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The fake home directory must have been created
	assert.DirExists(suite.T(), home)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	// The new environment must exist alongside the base environment
	assert.Contains(suite.T(), envs, "test")
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Empty(suite.T(), createdEnv.Mounts)
	assert.Nil(suite.T(), createdEnv.Command)
}

// TestCreateHyphenatedName tests that env create accepts a name containing
// hyphens and that the name round-trips through the YAML config file.
func (suite *EnvTestSuite) TestCreateHyphenatedName() {
	home := filepath.Join(tests.MoatTestDir, "create_hyphenated_name_home")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "my-env", "--home", home, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The fake home directory must have been created
	assert.DirExists(suite.T(), home)

	// The hyphenated environment name must be stored in the written config
	// file
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	// The new environment must exist alongside the base environment
	assert.Contains(suite.T(), envs, "test")
	createdEnv, exists := envs["my-env"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Empty(suite.T(), createdEnv.Mounts)
	assert.Nil(suite.T(), createdEnv.Command)
}

// TestCreateWithMounts tests that env create registers a new environment with
// a home directory and mount paths, and creates all of the referenced
// directories.
func (suite *EnvTestSuite) TestCreateWithMounts() {
	home := filepath.Join(tests.MoatTestDir, "create_mounts_home")
	mount1 := filepath.Join(tests.MoatTestDir, "create_mounts_mount1")
	mount2 := filepath.Join(tests.MoatTestDir, "create_mounts_mount2")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	require.NoError(suite.T(), utils.CleanupDir(mount1))
	require.NoError(suite.T(), utils.CleanupDir(mount2))
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(home))
		require.NoError(suite.T(), utils.CleanupDir(mount1))
		require.NoError(suite.T(), utils.CleanupDir(mount2))
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mount1, "--mount", mount2, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// All directories referenced by the environment must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), mount1)
	assert.DirExists(suite.T(), mount2)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	// The new environment must exist alongside the base environment
	assert.Contains(suite.T(), envs, "test")
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{mount1, mount2}, createdEnv.Mounts)
	assert.Nil(suite.T(), createdEnv.Command)
}

// TestCreateWithMountCWD tests that env create registers a new environment
// with working-directory mounting enabled when the --mountcwd flag is given
// with an empty value (empty means true).
func (suite *EnvTestSuite) TestCreateWithMountCWD() {
	home := filepath.Join(tests.MoatTestDir, "create_mountcwd_home")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--yes", "--mountcwd="})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The fake home directory must have been created
	assert.DirExists(suite.T(), home)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.NotNil(suite.T(), createdEnv.MountCWD)
	assert.True(suite.T(), *createdEnv.MountCWD)
}

// TestCreateWithMountCWDFalse tests that env create stores an explicit
// mountcwd: false when the --mountcwd flag is given the value false.
func (suite *EnvTestSuite) TestCreateWithMountCWDFalse() {
	home := filepath.Join(tests.MoatTestDir, "create_mountcwd_false_home")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	suite.T().Cleanup(func() { require.NoError(suite.T(), utils.CleanupDir(home)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--yes", "--mountcwd=false"})
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
	assert.False(suite.T(), *createdEnv.MountCWD)
}

// TestCreateWithMountDestinations tests that env create handles mounts given
// in the "source:destination" form: only the source directory is created and
// the mount is stored with the absolute source path.
func (suite *EnvTestSuite) TestCreateWithMountDestinations() {
	home := filepath.Join(tests.MoatTestDir, "create_mount_dests_home")
	mountSource := filepath.Join(tests.MoatTestDir, "create_mount_dests_src")
	mount := mountSource + ":/projects/src"
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	require.NoError(suite.T(), utils.CleanupDir(mountSource))
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(home))
		require.NoError(suite.T(), utils.CleanupDir(mountSource))
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mount, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The home and mount source directories must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), mountSource)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{mount}, createdEnv.Mounts)
}

// TestCreateWithRoMounts tests that env create registers a new environment
// with read-only mount paths and creates all of the referenced directories.
func (suite *EnvTestSuite) TestCreateWithRoMounts() {
	home := filepath.Join(tests.MoatTestDir, "create_ro_mounts_home")
	roMount1 := filepath.Join(tests.MoatTestDir, "create_ro_mounts_romount1")
	roMount2 := filepath.Join(tests.MoatTestDir, "create_ro_mounts_romount2")
	// Start from a clean slate so the test verifies directory creation
	require.NoError(suite.T(), utils.CleanupDir(home))
	require.NoError(suite.T(), utils.CleanupDir(roMount1))
	require.NoError(suite.T(), utils.CleanupDir(roMount2))
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(home))
		require.NoError(suite.T(), utils.CleanupDir(roMount1))
		require.NoError(suite.T(), utils.CleanupDir(roMount2))
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--ro-mount", roMount1, "--ro-mount", roMount2, "--yes"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// All directories referenced by the environment must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), roMount1)
	assert.DirExists(suite.T(), roMount2)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Empty(suite.T(), createdEnv.Mounts)
	assert.Equal(suite.T(), []string{roMount1, roMount2}, createdEnv.ReadOnlyMounts)
}

// TestCreateMountsPromptAccept tests that env create prompts for a missing
// mount source directory and creates it, storing the environment, when the
// user accepts the prompt.
func (suite *EnvTestSuite) TestCreateMountsPromptAccept() {
	home := filepath.Join(tests.MoatTestDir, "create_mounts_prompt_accept_home")
	mountSource := filepath.Join(tests.MoatTestDir, "create_mounts_prompt_accept_mount")
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(home))
		require.NoError(suite.T(), utils.CleanupDir(mountSource))
	})
	// Create the fake home directory beforehand so that only the mount
	// prompt is asked
	require.NoError(suite.T(), os.MkdirAll(home, 0755))
	tests.AnswerPrompt(suite.T(), "y")

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mountSource})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The mount source directory must have been created after the prompt
	assert.DirExists(suite.T(), mountSource)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), []string{mountSource}, createdEnv.Mounts)
}

// TestCreateMountsPromptDecline tests that env create aborts without
// creating the environment or the mount source directory when the user
// declines the prompt.
func (suite *EnvTestSuite) TestCreateMountsPromptDecline() {
	home := filepath.Join(tests.MoatTestDir, "create_mounts_prompt_decline_home")
	mountSource := filepath.Join(tests.MoatTestDir, "create_mounts_prompt_decline_mount")
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupDir(home))
		require.NoError(suite.T(), utils.CleanupDir(mountSource))
	})
	// Create the fake home directory beforehand so that only the mount
	// prompt is asked
	require.NoError(suite.T(), os.MkdirAll(home, 0755))
	tests.AnswerPrompt(suite.T(), "n")

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mountSource})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Aborting due to missing mount directories.")

	// Neither the environment nor the mount source directory was created
	assert.NoDirExists(suite.T(), mountSource)
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	assert.NotContains(suite.T(), envs, "newenv")
}
