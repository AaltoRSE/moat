package env_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/require"
)

// TestShow tests that the env show command prints the configuration of the
// named environment.
func (suite *EnvTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "show", "--name", "test"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Unmarshal output
	var outputEnv types.MoatEnv
	require.NoError(suite.T(), yaml.Unmarshal([]byte(capturedOutput), &outputEnv))
	assert.Equal(suite.T(), suite.BaseEnv, outputEnv)
}
