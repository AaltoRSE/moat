package runtime_test

import (
	"os"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// customRuntimeSpec returns a valid apptainer runtime specification for
// use in tests.
func customRuntimeSpec() types.RuntimeSpec {
	return types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:custom",
		CacheDir: tests.MoatTestDir + "/runtime_test_cache",
		PassEnv:  true,
	}
}

// TestList tests that runtime list prints the names of all available
// runtimes, including the default runtimes.
func (suite *RuntimeTestSuite) TestList() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "list"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer\n", capturedOutput)
}

// TestListWithName tests that runtime list with the -n/--name flag prints
// only the named runtime.
func (suite *RuntimeTestSuite) TestListWithName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "list", "--name", "apptainer"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer\n", capturedOutput)
}

// TestListWithMissingName tests that runtime list with the -n/--name flag
// given a name of a nonexistent runtime prints nothing.
func (suite *RuntimeTestSuite) TestListWithMissingName() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "list", "--name", "nonexistent"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), capturedOutput)
}

// TestListWithUserRuntime tests that runtime list also prints
// user-specified runtimes from the top-level runtimes key.
func (suite *RuntimeTestSuite) TestListWithUserRuntime() {

	configFile, _, err := tests.CreateTempConfig(nil, map[string]types.RuntimeSpec{"custom": customRuntimeSpec()})
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.Remove(configFile)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", configFile, "runtime", "list"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer\ncustom\n", capturedOutput)
}

// TestListWithOverwrittenDefaultRuntime tests that a user-specified
// runtime with the same name as a default runtime overwrites the default
// runtime.
func (suite *RuntimeTestSuite) TestListWithOverwrittenDefaultRuntime() {

	spec := customRuntimeSpec()
	configFile, _, err := tests.CreateTempConfig(nil, map[string]types.RuntimeSpec{"apptainer": spec})
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.Remove(configFile)) })

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", configFile, "runtime", "list"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	// Only one runtime with the name apptainer must be listed
	assert.Equal(suite.T(), "apptainer\n", capturedOutput)

	// The user-specified runtime spec must take precedence over the
	// default runtime spec
	cfg, err := config.InitConfig(configFile)
	require.NoError(suite.T(), err)
	runtimes := config.GetRuntimes(cfg)
	assert.Len(suite.T(), runtimes, 1)
	assert.Equal(suite.T(), spec.ImageUrl, runtimes["apptainer"].ImageUrl)
}
