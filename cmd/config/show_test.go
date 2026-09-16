package config_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

// TestShow tests that the config show command prints the full moat
// configuration, including the base environment.
func (suite *ConfigTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()

	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "show"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Unmarshal output
	var outputConfig types.Config
	require.NoError(suite.T(), yaml.Unmarshal([]byte(capturedOutput), &outputConfig))
	assert.Subset(suite.T(), outputConfig.Envs, map[string]types.MoatEnv{"test": suite.BaseEnv})
}
