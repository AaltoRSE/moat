package env_test

import (
	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
)

// TestList tests that env list without the name flag prints the names of all
// configured environments.
func (suite *EnvTestSuite) TestList() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Equal(suite.T(), "test\n", capturedOutput)
}

// TestListWithName tests that env list with the -n/--name flag prints only the
// named environment.
func (suite *EnvTestSuite) TestListWithName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list", "--name", "test"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Equal(suite.T(), "test\n", capturedOutput)
}

// TestListWithMissingName tests that env list with the -n/--name flag given a
// name of a nonexistent environment prints nothing.
func (suite *EnvTestSuite) TestListWithMissingName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "list", "--name", "nonexistent"})
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Empty(suite.T(), capturedOutput)
}
