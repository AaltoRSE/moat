package version_test

import (
	"testing"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/AaltoRSE/moat/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
	HomeDir string
}

func (suite *VersionTestSuite) SetupTest() {
	// Point the user home directory at an empty temporary directory so
	// that configuration initialization does not pick up a real
	// configuration file.
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)
}

// runVersion executes the moat version command with the given extra
// arguments and returns its standard output.
func (suite *VersionTestSuite) runVersion(args ...string) string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs(append([]string{"version"}, args...))
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	output, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	return output
}

// TestVersionPrintsMoatVersion tests that the version command prints the
// value of version.MoatVersion.
func (suite *VersionTestSuite) TestVersionPrintsMoatVersion() {
	// The version is normally set at build time via the linker flag
	// -X; set a known value for the duration of the test.
	originalVersion := version.MoatVersion
	version.MoatVersion = "v1.2.3-test"
	defer func() { version.MoatVersion = originalVersion }()

	output := suite.runVersion()

	assert.Contains(suite.T(), output, "v1.2.3-test")
}

// TestVersionInRootHelp tests that the version command is registered on
// the root command.
func (suite *VersionTestSuite) TestVersionInRootHelp() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--help"})
	require.NoError(suite.T(), rootCmd.Execute())
	output, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), output, "version")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
