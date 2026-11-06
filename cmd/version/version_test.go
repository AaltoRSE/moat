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
	require.NoError(suite.T(), rootCmd.Execute())
	output, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	return output
}

// setVersionMetadata sets the build-time version metadata to the given
// known values and restores the originals when the test finishes.
func (suite *VersionTestSuite) setVersionMetadata(v, os, arch string) {
	originalVersion, originalOS, originalArch := version.MoatVersion, version.MoatOS, version.MoatArch
	suite.T().Cleanup(func() {
		version.MoatVersion, version.MoatOS, version.MoatArch = originalVersion, originalOS, originalArch
	})
	version.MoatVersion, version.MoatOS, version.MoatArch = v, os, arch
}

// TestVersionPrintsMoatVersion tests that the version command prints the
// value of version.MoatVersion.
func (suite *VersionTestSuite) TestVersionPrintsMoatVersion() {
	// The version is normally set at build time via the linker flag
	// -X; set a known value for the duration of the test.
	suite.setVersionMetadata("v1.2.3-test", "", "")

	output := suite.runVersion()

	assert.Contains(suite.T(), output, "v1.2.3-test")
}

// TestVersionAllShortFlag tests that the version command prints the full
// version with the operating system and architecture when the -a flag is
// given.
func (suite *VersionTestSuite) TestVersionAllShortFlag() {
	suite.setVersionMetadata("v1.2.3-test", "linux", "amd64")

	output := suite.runVersion("-a")

	assert.Contains(suite.T(), output, "v1.2.3-test-Linux_x86_64")
}

// TestVersionAllLongFlag tests that the version command prints the full
// version with the operating system and architecture when the --all flag
// is given.
func (suite *VersionTestSuite) TestVersionAllLongFlag() {
	suite.setVersionMetadata("v1.2.3-test", "darwin", "arm64")

	output := suite.runVersion("--all")

	assert.Contains(suite.T(), output, "v1.2.3-test-Darwin_arm64")
}

// TestVersionAllFlagFallsBackToRuntimePlatform tests that the version
// command with the -a flag appends the platform of the system moat is
// running on when the binary was built without operating system and
// architecture information.
func (suite *VersionTestSuite) TestVersionAllFlagFallsBackToRuntimePlatform() {
	suite.setVersionMetadata("v1.2.3-test", "", "")

	output := suite.runVersion("-a")

	assert.Equal(suite.T(), "v1.2.3-test-"+version.PlatformName()+"\n", output)
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
