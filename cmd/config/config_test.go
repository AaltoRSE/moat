package config_test

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

type ConfigTestSuite struct {
	suite.Suite
	ConfigFile string
	BaseOutput string
	BaseEnv    types.MoatEnv
}

func (suite *ConfigTestSuite) SetupTest() {
	// Add environment called test to a fresh temporary moat-config
	suite.BaseEnv = types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	var err error
	suite.ConfigFile, suite.BaseOutput, err = tests.CreateTempConfig("test", suite.BaseEnv)
	if err != nil {
		panic(err)
	}
}

func (suite *ConfigTestSuite) TearDownTest() {
	// Remove the temporary config file
	err := os.Remove(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
}

// Test the show command
func (suite *ConfigTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()

	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "show"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Unmarshal output
	var outputConfig types.Config
	err = yaml.Unmarshal([]byte(capturedOutput), &outputConfig)
	if err != nil {
		panic(err)
	}
	assert.Subset(suite.T(), outputConfig.Envs, map[string]types.MoatEnv{"test": suite.BaseEnv})
}

// TestSet tests that config set updates the boolean mountcwd variable of an
// existing environment.
func (suite *ConfigTestSuite) TestSet() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "set", "envs.test.mountcwd", "true"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Set envs.test.mountcwd = true")

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

// TestGetConfigFile tests that GetConfigFile returns the active
// configuration file and falls back to the global moat-config.yaml when
// there is no active configuration.
func TestGetConfigFile(t *testing.T) {
	cfgFile, err := os.CreateTemp("", "moat-config.*.yaml")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(cfgFile.Name()))
	}()
	err = cfgFile.Close()
	assert.NoError(t, err)

	cfg, err := config.InitConfig(cfgFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, cfgFile.Name(), config.GetConfigFile(cfg))

	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "moat", "moat-config.yaml"), config.GetConfigFile(nil))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestShowTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
