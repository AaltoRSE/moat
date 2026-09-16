package runtimes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ApptainerRuntimeTestSuite tests the ApptainerRuntime against a fake
// apptainer binary that records its command line arguments to a log file
// instead of running a container.
type ApptainerRuntimeTestSuite struct {
	suite.Suite
	HomeDir  string
	WorkDir  string
	CacheDir string
	LogFile  string
}

// fakeApptainerScript is the body of the fake apptainer executable. The
// pull subcommand creates the target image file, and every other
// invocation (run) records its arguments, one per line, to the log file
// whose path is substituted for <logFile> when the script is written.
const fakeApptainerScript = `#!/bin/sh
if [ "$1" = "pull" ]; then
    : > "$2"
    exit 0
fi
printf '%s\n' "$@" > "<logFile>"
exit 0
`

func (suite *ApptainerRuntimeTestSuite) SetupTest() {
	// Isolate the home directory so that the home mount argument is
	// deterministic.
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)

	// Install a fake apptainer binary on the PATH.
	fakeBinDir := suite.T().TempDir()
	suite.LogFile = filepath.Join(fakeBinDir, "apptainer-args.log")
	script := strings.ReplaceAll(fakeApptainerScript, "<logFile>", suite.LogFile)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "apptainer"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Run the runtime from a dedicated working directory.
	suite.WorkDir = suite.T().TempDir()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), os.Chdir(suite.WorkDir))
	suite.T().Cleanup(func() { _ = os.Chdir(wd) })

	// Use a dedicated image cache directory.
	suite.CacheDir = suite.T().TempDir()
}

// newRuntime builds an ApptainerRuntime from a runtime spec with the given
// MountCWD setting.
func (suite *ApptainerRuntimeTestSuite) newRuntime(mountCWD bool) *ApptainerRuntime {
	return NewApptainerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "test/example-image:1.0",
		CacheDir: suite.CacheDir,
		PassEnv:  false,
		MountCWD: mountCWD,
	})
}

// runAndCaptureArgs executes the given command through the runtime and
// returns the arguments that the fake apptainer binary received.
func (suite *ApptainerRuntimeTestSuite) runAndCaptureArgs(runtime *ApptainerRuntime, env types.MoatEnv, args ...string) []string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	defer func() {
		_, _ = capture.StopCapture()
	}()

	code, err := runtime.Run(env, args, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, code)

	contents, err := os.ReadFile(suite.LogFile)
	require.NoError(suite.T(), err)
	return strings.Split(strings.TrimSpace(string(contents)), "\n")
}

// hasBindArg reports whether args contains a --bind flag whose value is
// exactly path.
func hasBindArg(args []string, path string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--bind" && args[i+1] == path {
			return true
		}
	}
	return false
}

// countBindArgs reports how many --bind flags in args have the value
// path.
func countBindArgs(args []string, path string) int {
	count := 0
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--bind" && args[i+1] == path {
			count++
		}
	}
	return count
}

// hasPwdArg reports whether args contains a --pwd flag whose value is
// exactly path.
func hasPwdArg(args []string, path string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--pwd" && args[i+1] == path {
			return true
		}
	}
	return false
}

// TestRunBaseArgs tests that the apptainer command line always starts with
// run --no-home --no-mount cwd, binds the fake home directory, and ends
// with the image path and the user command.
func (suite *ApptainerRuntimeTestSuite) TestRunBaseArgs() {
	home := filepath.Join(suite.HomeDir, "fake-home")
	require.NoError(suite.T(), os.MkdirAll(home, 0o755))
	env := types.MoatEnv{Home: home}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	// The base apptainer arguments must be present.
	assert.Equal(suite.T(), "run", apptainerArgs[0])
	assert.Contains(suite.T(), apptainerArgs, "--no-home")
	assert.Contains(suite.T(), apptainerArgs, "--no-mount")
	// The fake home directory must be bound to the host home directory.
	assert.True(suite.T(), hasBindArg(apptainerArgs, home+":"+suite.HomeDir))
	// The command line must end with the image path and the user command.
	assert.Equal(suite.T(), []string{"echo", "hello"}, apptainerArgs[len(apptainerArgs)-2:])
}

// TestRunDoesNotMountCWDByDefault tests that the working directory is not
// mounted when the runtime MountCWD setting is disabled and the
// environment does not set MountCWD.
func (suite *ApptainerRuntimeTestSuite) TestRunDoesNotMountCWDByDefault() {
	env := types.MoatEnv{Home: suite.HomeDir}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.False(suite.T(), hasBindArg(apptainerArgs, suite.WorkDir))
	assert.False(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// TestRunMountsCWDWhenRuntimeMountCWD tests that the working directory is
// mounted and used as the container working directory when the runtime
// MountCWD setting is enabled and the environment does not set MountCWD.
func (suite *ApptainerRuntimeTestSuite) TestRunMountsCWDWhenRuntimeMountCWD() {
	env := types.MoatEnv{Home: suite.HomeDir}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.True(suite.T(), hasBindArg(apptainerArgs, suite.WorkDir))
	assert.True(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDTrueOverridesRuntimeFalse tests that an environment
// MountCWD of true overrides a disabled runtime MountCWD setting.
func (suite *ApptainerRuntimeTestSuite) TestRunEnvMountCWDTrueOverridesRuntimeFalse() {
	mountCWD := true
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.True(suite.T(), hasBindArg(apptainerArgs, suite.WorkDir))
	assert.True(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDFalseOverridesRuntimeTrue tests that an environment
// MountCWD of false overrides an enabled runtime MountCWD setting.
func (suite *ApptainerRuntimeTestSuite) TestRunEnvMountCWDFalseOverridesRuntimeTrue() {
	mountCWD := false
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.False(suite.T(), hasBindArg(apptainerArgs, suite.WorkDir))
	assert.False(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDTrueWithRuntimeMountCWDTrue tests that the working
// directory is mounted when both the runtime MountCWD setting and the
// environment MountCWD are enabled.
func (suite *ApptainerRuntimeTestSuite) TestRunEnvMountCWDTrueWithRuntimeMountCWDTrue() {
	mountCWD := true
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.True(suite.T(), hasBindArg(apptainerArgs, suite.WorkDir))
	assert.True(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// TestRunDoesNotMountCWDWhenAlreadyMounted tests that the working
// directory is bound only once when it is already given as an environment
// mount, even when the runtime MountCWD setting is enabled.
func (suite *ApptainerRuntimeTestSuite) TestRunDoesNotMountCWDWhenAlreadyMounted() {
	env := types.MoatEnv{Home: suite.HomeDir, Mounts: []string{suite.WorkDir}}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.Equal(suite.T(), 1, countBindArgs(apptainerArgs, suite.WorkDir))
	assert.False(suite.T(), hasPwdArg(apptainerArgs, suite.WorkDir))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestApptainerRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(ApptainerRuntimeTestSuite))
}
