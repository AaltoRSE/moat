package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/AaltoRSE/moat/internal/logging"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	ConfigFile     *os.File
	ExpectedOutput string
	ExpectedEnv    types.MoatEnv
}

func (suite *ConfigTestSuite) SetupTest() {
	var (
		cfg *viper.Viper
		err error
	)

	suite.ConfigFile, err = os.CreateTemp("", "moat-config.*.yaml")
	if err != nil {
		panic(err)
	}
	err = suite.ConfigFile.Close()
	if err != nil {
		panic(err)
	}

	// Silence standard output during setup
	capture := utils.OutputCapture{}
	capture.StartCapture()
	logging.InitLogging(false)

	// Initialize configuration
	cfg, err = config.InitConfig(suite.ConfigFile.Name())
	if err != nil {
		panic(err)
	}

	// Add environment called test to moat-config
	suite.ExpectedEnv = types.MoatEnv{Home: "/tmp", Mounts: []string{}, ReadOnlyMounts: []string{}}
	err = env.CreateEnvironment(cfg, "test", suite.ExpectedEnv)
	if err != nil {
		panic(err)
	}

	// Capture config file contents
	configContents, err := os.ReadFile(suite.ConfigFile.Name())
	if err != nil {
		panic(err)
	}
	suite.ExpectedOutput = string(configContents)

	// Stop capturing output
	_, err = capture.StopCapture()
	fmt.Printf("%s", suite.ExpectedOutput)
	if err != nil {
		panic(err)
	}
}

func (suite *ConfigTestSuite) TearDownTest() {
	// Remove the temporary config file
	err := os.Remove(suite.ConfigFile.Name())
	if err != nil {
		panic(err)
	}
}

// Test the show command
func (suite *ConfigTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	cmd.RootCmd.SetArgs([]string{"--config", suite.ConfigFile.Name(), "config", "show"})
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

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestShowTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
