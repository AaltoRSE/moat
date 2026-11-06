package update_test

import (
	"testing"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UpdateCmdTestSuite tests the update command registration.
type UpdateCmdTestSuite struct {
	suite.Suite
	HomeDir string
}

func (suite *UpdateCmdTestSuite) SetupTest() {
	// Point the user home directory at an empty temporary directory so
	// that configuration initialization does not pick up a real
	// configuration file.
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)
}

// TestUpdateInRootHelp tests that the update command is registered on
// the root command.
func (suite *UpdateCmdTestSuite) TestUpdateInRootHelp() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--help"})
	require.NoError(suite.T(), rootCmd.Execute())
	output, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), output, "update")
}

// TestUpdateHelp tests that the update command shows its usage.
func (suite *UpdateCmdTestSuite) TestUpdateHelp() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"update", "--help"})
	require.NoError(suite.T(), rootCmd.Execute())
	output, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), output, "Update moat to the latest release")
}

// TestUpdateRejectsArguments tests that the update command rejects
// positional arguments.
func (suite *UpdateCmdTestSuite) TestUpdateRejectsArguments() {
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"update", "unexpected-argument"})

	err := rootCmd.Execute()

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "unknown command")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUpdateCmdTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateCmdTestSuite))
}
