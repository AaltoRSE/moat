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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "newenv", "--home", home, "--ro-mounts", roMount1, "--ro-mounts", roMount2, "--yes"})
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

// TestCopyWithRoMounts tests that env copy overrides the source environment's
// read-only mounts when the --ro-mounts flag is provided.
func (suite *EnvTestSuite) TestCopyWithRoMounts() {
	baseDir := filepath.Join(tests.MoatTestDir, "copy_ro_mounts")
	home := filepath.Join(baseDir, "home")
	roMount := filepath.Join(baseDir, "romount")
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv", "--home", home, "--ro-mounts", roMount, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The home and read-only mount source directories must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), roMount)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{roMount}, createdEnv.ReadOnlyMounts)
}

// TestSetCommand tests that env set updates the command variable of an
// existing environment.
func (suite *EnvTestSuite) TestSetCommand() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "command", "code --wait ."})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.command = code --wait .")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), suite.BaseEnv.Home, updatedEnv.Home)
	assert.NotNil(suite.T(), updatedEnv.Command)
	assert.Equal(suite.T(), "code --wait .", *updatedEnv.Command)
}

// TestSetMounts tests that env set replaces the mounts variable of an
// existing environment with the given list of mount paths.
func (suite *EnvTestSuite) TestSetMounts() {
	mount1 := filepath.Join(tests.MoatTestDir, "set_mounts_1")
	mount2 := filepath.Join(tests.MoatTestDir, "set_mounts_2")
	// The mount source directories must exist for the configuration to be valid
	err := os.MkdirAll(mount1, 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(mount2, 0755)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(mount1)
		_ = os.RemoveAll(mount2)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "mounts", mount1, mount2})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.mounts = [")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), []string{mount1, mount2}, updatedEnv.Mounts)
}

// TestSetMountCWD tests that env set updates the boolean mountcwd variable of
// an existing environment.
func (suite *EnvTestSuite) TestSetMountCWD() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "mountcwd", "true"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.mountcwd = true")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), updatedEnv.MountCWD)
	assert.True(suite.T(), *updatedEnv.MountCWD)
}

// TestSetHome tests that env set updates the home variable of an existing
// environment, which is already set in the configuration.
func (suite *EnvTestSuite) TestSetHome() {
	newHome := filepath.Join(tests.MoatTestDir, "set_home")
	// The new home directory must exist for the configuration to be valid
	err := os.MkdirAll(newHome, 0755)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(newHome)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "home", newHome})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.home = "+newHome)

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), newHome, updatedEnv.Home)
}

// TestSetLeavesOtherEnvsUnchanged tests that env set only modifies the
// environment given by the --name flag and leaves other environments
// untouched.
func (suite *EnvTestSuite) TestSetLeavesOtherEnvsUnchanged() {
	otherHome := filepath.Join(tests.MoatTestDir, "set_other_home")
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(otherHome)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(otherHome)
	})

	// Create a second environment to verify it stays untouched
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "create", "--name", "other", "--home", otherHome, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	_, err = capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Set a variable on the "test" environment
	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "command", "opencode"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	_, err = capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Unmarshal the updated configuration and check both environments
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), updatedEnv.Command)
	assert.Equal(suite.T(), "opencode", *updatedEnv.Command)
	otherEnv, exists := envs["other"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), otherHome, otherEnv.Home)
	assert.Nil(suite.T(), otherEnv.Command)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvTestSuite(t *testing.T) {
	suite.Run(t, new(EnvTestSuite))
}
