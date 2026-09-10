package config_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
)

// TestShow tests that the config show command prints the full moat
// configuration, including the base environment.
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
