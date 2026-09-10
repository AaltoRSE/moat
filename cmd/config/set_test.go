package config_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
)

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
	assert.Contains(suite.T(), capturedOutput, "Set envs.test.mountcwd = [true]")

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
