package run_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	types "github.com/AaltoRSE/moat/internal/types"
)

// RunTestSuite tests the run command: the pre-flight check that creates
// missing mount source directories, with and without prompting.
type RunTestSuite struct {
	suite.Suite
	ConfigFile  string
	Mount       string
	MountSource string
}

func (suite *RunTestSuite) SetupTest() {
	// Add an environment called test with a mount whose source directory
	// does not exist and without a default command to a fresh temporary
	// moat-config. CreateTempConfig creates the mount directory, so each
	// test removes it first to verify its creation during the run.
	suite.MountSource = filepath.Join(tests.MoatTestDir, "run_missing_mount")
	suite.Mount = suite.MountSource + ":/projects/missing"
	env := types.MoatEnv{
		Home:   tests.MoatTestDir,
		Mounts: []string{suite.Mount},
	}
	var err error
	suite.ConfigFile, _, err = tests.CreateTempConfig(map[string]types.MoatEnv{"test": env}, nil)
	require.NoError(suite.T(), err)
}

func (suite *RunTestSuite) TearDownTest() {
	// Remove the temporary config file
	suite.Require().NoError(os.Remove(suite.ConfigFile))
}

// run executes moat run for the test environment.
func (suite *RunTestSuite) run(extraArgs ...string) string {
	args := append([]string{"--config", suite.ConfigFile, "run", "--name", "test"}, extraArgs...)
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs(args)
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	return capturedOutput
}

// TestRunCreatesMissingMountsWithYes tests that run creates the missing
// mount source directories of the environment without prompting when --yes
// is given. Without a command, run stops after the pre-flight check.
func (suite *RunTestSuite) TestRunCreatesMissingMountsWithYes() {
	require.NoError(suite.T(), os.RemoveAll(suite.MountSource))

	suite.run("--yes")

	assert.DirExists(suite.T(), suite.MountSource)
}

// TestRunCreatesMissingMountsOnPromptAccept tests that run asks whether a
// missing mount source directory should be created and creates it when the
// user accepts.
func (suite *RunTestSuite) TestRunCreatesMissingMountsOnPromptAccept() {
	require.NoError(suite.T(), os.RemoveAll(suite.MountSource))
	tests.AnswerPrompt(suite.T(), "y")

	suite.run()

	assert.DirExists(suite.T(), suite.MountSource)
}

// TestRunAbortsOnPromptDecline tests that run does not create a missing
// mount source directory and aborts when the user declines the prompt.
func (suite *RunTestSuite) TestRunAbortsOnPromptDecline() {
	require.NoError(suite.T(), os.RemoveAll(suite.MountSource))
	tests.AnswerPrompt(suite.T(), "n")

	capturedOutput := suite.run()

	assert.NoDirExists(suite.T(), suite.MountSource)
	assert.Contains(suite.T(), capturedOutput, "Aborting due to missing mount directories.")
}

// TestRunCreatesMissingHomeWithYes tests that run creates the missing fake
// home directory of the environment without prompting when --yes is given.
func (suite *RunTestSuite) TestRunCreatesMissingHomeWithYes() {
	home := filepath.Join(tests.MoatTestDir, "run_missing_home")
	env := types.MoatEnv{Home: home}
	configFile, _, err := tests.CreateTempConfig(map[string]types.MoatEnv{"homeenv": env}, nil)
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() { _ = os.Remove(configFile) })
	// CreateTempConfig creates the fake home directory; remove it so the
	// test verifies its creation during the run
	require.NoError(suite.T(), os.RemoveAll(home))

	args := []string{"--config", configFile, "run", "--name", "homeenv", "--yes"}
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs(args)
	require.NoError(suite.T(), rootCmd.Execute())
	_, err = capture.StopCapture()
	require.NoError(suite.T(), err)

	assert.DirExists(suite.T(), home)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRunTestSuite(t *testing.T) {
	suite.Run(t, new(RunTestSuite))
}
