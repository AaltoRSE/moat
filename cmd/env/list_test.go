package env_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestList tests that env list without the name flag prints the names of all
// configured environments.
func (suite *EnvTestSuite) TestList() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test\n", capturedOutput)
}

// TestListWithName tests that env list with the -n/--name flag prints only the
// named environment.
func (suite *EnvTestSuite) TestListWithName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list", "--name", "test"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test\n", capturedOutput)
}

// TestListWithMissingName tests that env list with the -n/--name flag given a
// name of a nonexistent environment prints nothing.
func (suite *EnvTestSuite) TestListWithMissingName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list", "--name", "nonexistent"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), capturedOutput)
}
