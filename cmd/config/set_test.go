package config_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSet tests that config set updates the boolean mountcwd variable of an
// existing environment.
func (suite *ConfigTestSuite) TestSet() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "config", "set", "envs.test.mountcwd", "true"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set envs.test.mountcwd = [true]")

	// Unmarshal the updated configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	envs := config.GetEnvs(cfg)
	updatedEnv, exists := envs["test"]
	assert.True(suite.T(), exists)
	assert.NotNil(suite.T(), updatedEnv.MountCWD)
	assert.True(suite.T(), *updatedEnv.MountCWD)
}
