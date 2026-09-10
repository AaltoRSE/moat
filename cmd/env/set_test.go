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

// TestSetCommand tests that env set updates the command variable of an
// existing environment using the --command flag.
func (suite *EnvTestSuite) TestSetCommand() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--command", "code --wait ."})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.command = [code --wait .]")

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
// existing environment with the list of mount paths given by repeated
// --mount flags.
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--mount", mount1, "--mount", mount2})
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

// TestSetRoMounts tests that env set replaces the readonlymounts variable of
// an existing environment with the list of paths given by repeated --ro-mount
// flags.
func (suite *EnvTestSuite) TestSetRoMounts() {
	roMount1 := filepath.Join(tests.MoatTestDir, "set_romounts_1")
	roMount2 := filepath.Join(tests.MoatTestDir, "set_romounts_2")
	// The read-only mount source directories must exist for the configuration to be valid
	err := os.MkdirAll(roMount1, 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(roMount2, 0755)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(roMount1)
		_ = os.RemoveAll(roMount2)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--ro-mount", roMount1, "--ro-mount", roMount2})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.readonlymounts = [")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), []string{roMount1, roMount2}, updatedEnv.ReadOnlyMounts)
}

// TestSetHome tests that env set updates the home variable of an existing
// environment, which is already set in the configuration, using the --home
// flag.
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--home", newHome})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.home = ["+newHome+"]")

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

// TestSetMultipleFlags tests that env set applies several given flags in a
// single invocation and leaves the other variables of the environment
// untouched.
func (suite *EnvTestSuite) TestSetMultipleFlags() {
	newHome := filepath.Join(tests.MoatTestDir, "set_multiple_home")
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--home", newHome, "--command", "opencode"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set test.home = ["+newHome+"]")
	assert.Contains(suite.T(), capturedOutput, "Set test.command = [opencode]")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), newHome, updatedEnv.Home)
	assert.NotNil(suite.T(), updatedEnv.Command)
	assert.Equal(suite.T(), "opencode", *updatedEnv.Command)
	assert.Empty(suite.T(), updatedEnv.Mounts)
	assert.Empty(suite.T(), updatedEnv.ReadOnlyMounts)
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
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "set", "--name", "test", "--command", "opencode"})
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
