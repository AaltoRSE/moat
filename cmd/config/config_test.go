package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/cmd"
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
	ConfigFile     string
	ExpectedOutput string
	ExpectedEnv    types.MoatEnv
}

func (suite *ConfigTestSuite) SetupTest() {
	// Add environment called test to a fresh temporary moat-config
	suite.ExpectedEnv = types.MoatEnv{Home: "/tmp", Mounts: []string{}, ReadOnlyMounts: []string{}, Command: ""}
	var err error
	suite.ConfigFile, suite.ExpectedOutput, err = tests.CreateTempConfig("test", suite.ExpectedEnv)
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
	cmd.RootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "show"})
	err := cmd.RootCmd.Execute()
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
	assert.Subset(suite.T(), outputConfig.Envs, map[string]types.MoatEnv{"test": suite.ExpectedEnv})
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
