package env_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
)

// TestShow tests that the env show command prints the configuration of the
// named environment.
func (suite *EnvTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "show", "--name", "test"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Unmarshal output
	var outputEnv types.MoatEnv
	err = yaml.Unmarshal([]byte(capturedOutput), &outputEnv)
	if err != nil {
		panic(err)
	}
	assert.Equal(suite.T(), suite.BaseEnv, outputEnv)
}
