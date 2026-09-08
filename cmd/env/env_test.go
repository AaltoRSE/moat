package env_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/suite"
)

type EnvTestSuite struct {
	suite.Suite
	ConfigFile string
	BaseOutput string
	BaseEnv    types.MoatEnv
}

func (suite *EnvTestSuite) SetupTest() {
	// Add environment called test to a fresh temporary moat-config
	suite.BaseEnv = types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	var err error
	suite.ConfigFile, suite.BaseOutput, err = tests.CreateTempConfig("test", suite.BaseEnv)
	if err != nil {
		panic(err)
	}
}

func (suite *EnvTestSuite) TearDownTest() {
	// Remove the temporary config file
	err := os.Remove(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
}

// TestShow tests that the env show command prints the configuration of the
// named environment.
func (suite *EnvTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "show", "--name", "test"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Unmarshal output
	var outputEnv types.MoatEnv
	err = yaml.Unmarshal([]byte(capturedOutput), &outputEnv)
	if err != nil {
		panic(err)
	}
	assert.Equal(suite.T(), suite.BaseEnv, outputEnv)
}

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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mounts", mount1, "--mounts", mount2, "--yes"})
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--mounts", mount, "--yes"})
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvTestSuite(t *testing.T) {
	suite.Run(t, new(EnvTestSuite))
}
