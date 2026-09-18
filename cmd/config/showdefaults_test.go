package config_test

import (
	"strings"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

// TestShowDefaults tests that the config show-defaults command prints the
// default moat configuration, independent of the loaded configuration
// file.
func (suite *ConfigTestSuite) TestShowDefaults() {

	capture := utils.OutputCapture{}
	capture.StartCapture()

	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "show-defaults"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// The output must match the default configuration exactly.
	assert.Equal(suite.T(), config.GetConfigAsString(config.CreateDefaultConfig()), strings.TrimSuffix(capturedOutput, "\n"))

	// The output must not contain the base environment from the loaded
	// configuration.
	var outputConfig types.Config
	require.NoError(suite.T(), yaml.Unmarshal([]byte(capturedOutput), &outputConfig))
	assert.Empty(suite.T(), outputConfig.Envs)
}
