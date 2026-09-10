package env_test

import (
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
)

// TestCreateHomeOnly tests that env create registers a new environment with
// only a home directory (no mounts) and creates the fake home directory.
func (suite *EnvTestSuite) TestCreateHomeOnly() {
	baseDir := filepath.Join(tests.MoatTestDir, "create_home_only")
	home := filepath.Join(baseDir, "home")
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The fake home directory must have been created
	assert.DirExists(suite.T(), home)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	// The new environment must exist alongside the base environment
	assert.Contains(suite.T(), envs, "test")
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Empty(suite.T(), createdEnv.Mounts)
	assert.Nil(suite.T(), createdEnv.Command)
}

// TestCreateWithMounts tests that env create registers a new environment with
// a home directory and mount paths, and creates all of the referenced
// directories.
func (suite *EnvTestSuite) TestCreateWithMounts() {
	baseDir := filepath.Join(tests.MoatTestDir, "create_mounts")
	home := filepath.Join(baseDir, "home")
	mount1 := filepath.Join(baseDir, "mount1")
	mount2 := filepath.Join(baseDir, "mount2")
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mount1, "--mount", mount2, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// All directories referenced by the environment must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), mount1)
	assert.DirExists(suite.T(), mount2)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	// The new environment must exist alongside the base environment
	assert.Contains(suite.T(), envs, "test")
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{mount1, mount2}, createdEnv.Mounts)
	assert.Nil(suite.T(), createdEnv.Command)
}

// TestCreateWithMountDestinations tests that env create handles mounts given
// in the "source:destination" form: only the source directory is created and
// the mount is stored with the absolute source path.
func (suite *EnvTestSuite) TestCreateWithMountDestinations() {
	baseDir := filepath.Join(tests.MoatTestDir, "create_mount_dests")
	home := filepath.Join(baseDir, "home")
	mountSource := filepath.Join(baseDir, "src")
	mount := mountSource + ":/projects/src"
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mount", mount, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The home and mount source directories must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), mountSource)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{mount}, createdEnv.Mounts)
}

// TestCreateWithRoMounts tests that env create registers a new environment
// with read-only mount paths and creates all of the referenced directories.
func (suite *EnvTestSuite) TestCreateWithRoMounts() {
	baseDir := filepath.Join(tests.MoatTestDir, "create_ro_mounts")
	home := filepath.Join(baseDir, "home")
	roMount1 := filepath.Join(baseDir, "romount1")
	roMount2 := filepath.Join(baseDir, "romount2")
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--ro-mount", roMount1, "--ro-mount", roMount2, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// All directories referenced by the environment must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), roMount1)
	assert.DirExists(suite.T(), roMount2)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Empty(suite.T(), createdEnv.Mounts)
	assert.Equal(suite.T(), []string{roMount1, roMount2}, createdEnv.ReadOnlyMounts)
}
