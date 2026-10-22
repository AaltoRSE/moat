package engines

import (
	"fmt"
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

// DockerRuntimeTestSuite tests the DockerRuntime against a fake docker
// binary that records its command line arguments to a log file instead of
// running containers.
type DockerRuntimeTestSuite struct {
	suite.Suite
	HomeDir        string
	WorkDir        string
	CacheDir       string
	LogFile        string
	PasswdDumpFile string
}

// fakeDockerScript is the body of the fake docker executable. It handles
// the invocations made by [DockerRuntime]:
//
//  1. pull <image>: the pull is logged and succeeds, unless the image
//     reference contains "fail", in which case the pull fails.
//  2. run ... <image> cp /etc/passwd <file>: a minimal passwd file is
//     written to <file>, as a container would deliver it; the invocation
//     is not logged.
//  3. run ... <image> cp /etc/group <file>: a minimal group file is
//     written to <file>, as a container would deliver it; the invocation
//     is not logged.
//  4. run ... <image> <args>: the arguments are logged, the contents of a
//     bind-mounted passwd file are appended to the passwd dump file, and
//     the run succeeds.
//  5. images --format <format> <repository>: the references stored for
//     the repository are printed — two tags for test/example-image and
//     one for localhost:5000/moat, nothing for repositories containing
//     "none"; the invocation is not logged.
//
// The placeholders <logFile> and <passwdDumpFile> are substituted with the
// log and dump file paths when the script is written.
const fakeDockerScript = `#!/bin/sh
PATH=/usr/bin:/bin
export PATH

# The passwd and group copy runs carry /etc/passwd or /etc/group as the
# second-to-last argument: write a minimal file to the target and exit
# without logging.
prev=""
for arg in "$@"; do
    if [ "$prev" = "/etc/passwd" ]; then
        printf 'root:x:0:0:root:/root:/bin/sh\n' > "$arg"
        exit 0
    fi
    if [ "$prev" = "/etc/group" ]; then
        printf 'root:x:0:\n' > "$arg"
        exit 0
    fi
    prev="$arg"
done

if [ "$1" = "images" ]; then
    case "$4" in
    *none*)
        exit 0
        ;;
    test/example-image)
        printf 'test/example-image:1.1\n'
        printf 'test/example-image:1.0\n'
        exit 0
        ;;
    localhost:5000/moat)
        printf 'localhost:5000/moat:latest\n'
        exit 0
        ;;
    esac
    exit 0
fi

if [ "$1" = "pull" ]; then
    case "$2" in
    *fail*)
        exit 1
        ;;
    esac
fi

# Log the arguments of the invocation, followed by an empty line.
printf '%s\n' "$@" >> "<logFile>"
printf '\n' >> "<logFile>"

# Dump the contents of a bind-mounted passwd file, if any.
for arg in "$@"; do
    case "$arg" in
    type=bind,source=*,target=/etc/passwd,readonly)
        src="${arg#type=bind,source=}"
        src="${src%,target=/etc/passwd,readonly}"
        cat "$src" >> "<passwdDumpFile>" 2>/dev/null
        ;;
    esac
done

exit 0
`

func (suite *DockerRuntimeTestSuite) SetupTest() {
	// Isolate the home directory and the user name so that the home mount
	// argument and the appended passwd line are deterministic.
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)
	suite.T().Setenv("USER", "moattest")

	// Install a fake docker binary on the PATH.
	fakeBinDir := suite.T().TempDir()
	suite.LogFile = filepath.Join(fakeBinDir, "docker-args.log")
	suite.PasswdDumpFile = filepath.Join(fakeBinDir, "docker-passwd.dump")
	script := fakeDockerScript
	script = strings.ReplaceAll(script, "<logFile>", suite.LogFile)
	script = strings.ReplaceAll(script, "<passwdDumpFile>", suite.PasswdDumpFile)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "docker"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Run the runtime from a dedicated working directory.
	suite.WorkDir = suite.T().TempDir()
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), os.Chdir(suite.WorkDir))
	suite.T().Cleanup(func() { _ = os.Chdir(wd) })

	// Use a dedicated scratch directory for the temporary passwd file. The
	// Docker runtime does not create it, so it must exist beforehand.
	suite.CacheDir = suite.T().TempDir()
}

// newRuntime builds a DockerRuntime from a runtime spec with the given
// MountCWD setting.
func (suite *DockerRuntimeTestSuite) newRuntime(mountCWD bool) *DockerRuntime {
	return NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "test/example-image:1.0",
		CacheDir: suite.CacheDir,
		PassEnv:  false,
		MountCWD: mountCWD,
	})
}

// cwdMountValue returns the value of the --mount flag that bind-mounts
// the working directory at its own path.
func (suite *DockerRuntimeTestSuite) cwdMountValue() string {
	return fmt.Sprintf("type=bind,source=%s,target=%s", suite.WorkDir, suite.WorkDir)
}

// readLogBlocks reads the fake docker log file and returns its blocks: a
// block of arguments per logged invocation, split on the empty lines that
// separate the invocations.
func (suite *DockerRuntimeTestSuite) readLogBlocks() [][]string {
	contents, err := os.ReadFile(suite.LogFile)
	require.NoError(suite.T(), err)
	var blocks [][]string
	for _, block := range strings.Split(string(contents), "\n\n") {
		if block = strings.TrimSpace(block); block != "" {
			blocks = append(blocks, strings.Split(block, "\n"))
		}
	}
	return blocks
}

// runAndCaptureArgs executes the given command through the runtime and
// returns the arguments that the fake docker binary received for the
// container run.
func (suite *DockerRuntimeTestSuite) runAndCaptureArgs(runtime *DockerRuntime, env types.MoatEnv, args ...string) []string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	defer func() {
		_, _ = capture.StopCapture()
	}()

	code, err := runtime.Run(env, args, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, code)

	blocks := suite.readLogBlocks()
	require.NotEmpty(suite.T(), blocks)
	return blocks[len(blocks)-1]
}

// hasMountArg reports whether args contains a --mount flag whose value is
// exactly value.
func hasMountArg(args []string, value string) bool {
	return countMountArgs(args, value) > 0
}

// countMountArgs reports how many --mount flags in args have the value
// value.
func countMountArgs(args []string, value string) int {
	count := 0
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--mount" && args[i+1] == value {
			count++
		}
	}
	return count
}

// hasWorkdirArg reports whether args contains a --workdir flag whose
// value is exactly path.
func hasWorkdirArg(args []string, path string) bool {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--workdir" && args[i+1] == path {
			return true
		}
	}
	return false
}

// argValueAfter returns the argument that directly follows flag in args,
// or an empty string when flag is not present.
func argValueAfter(args []string, flag string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

// passwdMountArg returns the value of the --mount flag that bind-mounts
// a temporary passwd file from cacheDir read-only at /etc/passwd, or an
// empty string when no such mount is present.
func passwdMountArg(args []string, cacheDir string) string {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != "--mount" {
			continue
		}
		value := args[i+1]
		if strings.HasPrefix(value, "type=bind,source="+cacheDir+string(os.PathSeparator)+"passwd.") &&
			strings.HasSuffix(value, ",target=/etc/passwd,readonly") {
			return value
		}
	}
	return ""
}

// TestRunBaseArgs tests that the docker command line always starts with
// run --read-only --rm --tty --interactive, runs as the current user in a
// container named after moat, binds the fake home directory to the host
// home directory, mounts a passwd file from the scratch directory, and
// ends with the image reference and the user command.
func (suite *DockerRuntimeTestSuite) TestRunBaseArgs() {
	home := filepath.Join(suite.HomeDir, "fake-home")
	require.NoError(suite.T(), os.MkdirAll(home, 0o755))
	env := types.MoatEnv{Home: home}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	// The base docker arguments must be present.
	assert.Equal(suite.T(), "run", dockerArgs[0])
	assert.Contains(suite.T(), dockerArgs, "--read-only")
	assert.Contains(suite.T(), dockerArgs, "--rm")
	assert.Contains(suite.T(), dockerArgs, "--tty")
	assert.Contains(suite.T(), dockerArgs, "--interactive")

	// The container must run as the current user.
	assert.Contains(suite.T(), dockerArgs, "--user")
	assert.Contains(suite.T(), dockerArgs, fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()))

	// The container must get a generated moat- name and the same hostname.
	name := argValueAfter(dockerArgs, "--name")
	assert.Regexp(suite.T(), `^moat-[a-z0-9]{4}$`, name)
	assert.Equal(suite.T(), name, argValueAfter(dockerArgs, "--hostname"))

	// The fake home directory must be mounted at the host home directory.
	assert.True(suite.T(), hasMountArg(dockerArgs, fmt.Sprintf("type=bind,source=%s,target=%s", home, suite.HomeDir)))

	// A temporary passwd file from the scratch directory must be mounted
	// at /etc/passwd.
	assert.NotEmpty(suite.T(), passwdMountArg(dockerArgs, suite.CacheDir))

	// The command line must end with the image reference and the user
	// command.
	assert.Equal(suite.T(), []string{"test/example-image:1.0", "echo", "hello"}, dockerArgs[len(dockerArgs)-3:])
}

// TestRunMounts tests that read-write mounts are passed to docker as bind
// mounts, at the same path or at the destination given in the mount
// specification.
func (suite *DockerRuntimeTestSuite) TestRunMounts() {
	env := types.MoatEnv{
		Home:   suite.HomeDir,
		Mounts: []string{"/data/project", "/data/docs:/usr/share/doc"},
	}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.True(suite.T(), hasMountArg(dockerArgs, "type=bind,source=/data/project,target=/data/project"))
	assert.True(suite.T(), hasMountArg(dockerArgs, "type=bind,source=/data/docs,target=/usr/share/doc"))
}

// TestRunReadOnlyMounts tests that read-only mounts are passed to docker
// as read-only bind mounts.
func (suite *DockerRuntimeTestSuite) TestRunReadOnlyMounts() {
	env := types.MoatEnv{
		Home:           suite.HomeDir,
		ReadOnlyMounts: []string{"/data/tools", "/data/docs:/usr/share/doc"},
	}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.True(suite.T(), hasMountArg(dockerArgs, "type=bind,source=/data/tools,target=/data/tools,readonly"))
	assert.True(suite.T(), hasMountArg(dockerArgs, "type=bind,source=/data/docs,target=/usr/share/doc,readonly"))
}

// TestRunDoesNotMountCWDByDefault tests that the working directory is not
// mounted when the runtime MountCWD setting is disabled and the
// environment does not set MountCWD.
func (suite *DockerRuntimeTestSuite) TestRunDoesNotMountCWDByDefault() {
	env := types.MoatEnv{Home: suite.HomeDir}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.False(suite.T(), hasMountArg(dockerArgs, suite.cwdMountValue()))
	assert.False(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunMountsCWDWhenRuntimeMountCWD tests that the working directory is
// mounted and used as the container working directory when the runtime
// MountCWD setting is enabled and the environment does not set MountCWD.
func (suite *DockerRuntimeTestSuite) TestRunMountsCWDWhenRuntimeMountCWD() {
	env := types.MoatEnv{Home: suite.HomeDir}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.True(suite.T(), hasMountArg(dockerArgs, suite.cwdMountValue()))
	assert.True(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDTrueOverridesRuntimeFalse tests that an environment
// MountCWD of true overrides a disabled runtime MountCWD setting.
func (suite *DockerRuntimeTestSuite) TestRunEnvMountCWDTrueOverridesRuntimeFalse() {
	mountCWD := true
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	assert.True(suite.T(), hasMountArg(dockerArgs, suite.cwdMountValue()))
	assert.True(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDFalseOverridesRuntimeTrue tests that an environment
// MountCWD of false overrides an enabled runtime MountCWD setting.
func (suite *DockerRuntimeTestSuite) TestRunEnvMountCWDFalseOverridesRuntimeTrue() {
	mountCWD := false
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.False(suite.T(), hasMountArg(dockerArgs, suite.cwdMountValue()))
	assert.False(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunEnvMountCWDTrueWithRuntimeMountCWDTrue tests that the working
// directory is mounted when both the runtime MountCWD setting and the
// environment MountCWD are enabled.
func (suite *DockerRuntimeTestSuite) TestRunEnvMountCWDTrueWithRuntimeMountCWDTrue() {
	mountCWD := true
	env := types.MoatEnv{Home: suite.HomeDir, MountCWD: &mountCWD}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.True(suite.T(), hasMountArg(dockerArgs, suite.cwdMountValue()))
	assert.True(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunDoesNotMountCWDWhenAlreadyMounted tests that the working
// directory is mounted only once when it is already given as an
// environment mount, even when the runtime MountCWD setting is enabled.
func (suite *DockerRuntimeTestSuite) TestRunDoesNotMountCWDWhenAlreadyMounted() {
	env := types.MoatEnv{Home: suite.HomeDir, Mounts: []string{suite.WorkDir}}

	dockerArgs := suite.runAndCaptureArgs(suite.newRuntime(true), env, "echo", "hello")

	assert.Equal(suite.T(), 1, countMountArgs(dockerArgs, suite.cwdMountValue()))
	assert.False(suite.T(), hasWorkdirArg(dockerArgs, suite.WorkDir))
}

// TestRunAppendsCurrentUserToPasswdFile tests that the mounted passwd
// file contains the account line of the current user, appended to the
// passwd file delivered by the image.
func (suite *DockerRuntimeTestSuite) TestRunAppendsCurrentUserToPasswdFile() {
	env := types.MoatEnv{Home: suite.HomeDir}

	suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	contents, err := os.ReadFile(suite.PasswdDumpFile)
	require.NoError(suite.T(), err)
	expected := "root:x:0:0:root:/root:/bin/sh\n" +
		fmt.Sprintf("%s:x:%d:%d::%s:/bin/bash\n", "moattest", os.Getuid(), os.Getgid(), suite.HomeDir)
	assert.Equal(suite.T(), expected, string(contents))
}

// TestRunNoCommand tests that Run reports an error and does not invoke
// docker when no command is given.
func (suite *DockerRuntimeTestSuite) TestRunNoCommand() {
	env := types.MoatEnv{Home: suite.HomeDir}

	code, err := suite.newRuntime(false).Run(env, nil, nil)

	assert.Equal(suite.T(), 1, code)
	require.Error(suite.T(), err)
	require.NoFileExists(suite.T(), suite.LogFile)
}

// TestPull tests that Pull pulls the image with docker and returns the
// image reference.
func (suite *DockerRuntimeTestSuite) TestPull() {
	imageUrl, err := suite.newRuntime(false).Pull(false)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test/example-image:1.0", imageUrl)

	blocks := suite.readLogBlocks()
	require.Len(suite.T(), blocks, 1)
	assert.Equal(suite.T(), []string{"pull", "test/example-image:1.0"}, blocks[0])
}

// TestPullFailsWhenDockerPullFails tests that Pull reports the error of a
// failing docker pull.
func (suite *DockerRuntimeTestSuite) TestPullFailsWhenDockerPullFails() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "fail/example-image:1.0",
		CacheDir: suite.CacheDir,
		PassEnv:  false,
	})

	_, err := runtime.Pull(false)

	require.Error(suite.T(), err)
}

// TestRunFailsWhenPullFails tests that Run reports an error and does not
// run the command when the image cannot be pulled.
func (suite *DockerRuntimeTestSuite) TestRunFailsWhenPullFails() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "fail/example-image:1.0",
		CacheDir: suite.CacheDir,
		PassEnv:  false,
	})
	env := types.MoatEnv{Home: suite.HomeDir}

	code, err := runtime.Run(env, []string{"echo", "hello"}, nil)

	assert.Equal(suite.T(), 1, code)
	require.Error(suite.T(), err)
	// No successful docker invocation may have been logged.
	require.NoFileExists(suite.T(), suite.LogFile)
}

// TestNewRuntimeDefaultsExecutable tests that the runtime defaults to the
// docker executable when the runtime spec does not set one.
func (suite *DockerRuntimeTestSuite) TestNewRuntimeDefaultsExecutable() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{Type: "docker"})

	assert.Equal(suite.T(), "docker", runtime.Executable)
}

// TestNewRuntimeUsesSpecExecutable tests that the runtime uses the
// executable given in the runtime spec.
func (suite *DockerRuntimeTestSuite) TestNewRuntimeUsesSpecExecutable() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{Type: "docker", Executable: "my-docker"})

	assert.Equal(suite.T(), "my-docker", runtime.Executable)
}

// TestRunUsesSpecExecutable tests that Run invokes the executable given
// in the runtime spec instead of the default docker binary.
func (suite *DockerRuntimeTestSuite) TestRunUsesSpecExecutable() {
	// Install a fake binary under a custom name on the PATH.
	fakeBinDir := suite.T().TempDir()
	customLogFile := filepath.Join(fakeBinDir, "custom-args.log")
	customPasswdDumpFile := filepath.Join(fakeBinDir, "custom-passwd.dump")
	script := fakeDockerScript
	script = strings.ReplaceAll(script, "<logFile>", customLogFile)
	script = strings.ReplaceAll(script, "<passwdDumpFile>", customPasswdDumpFile)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "my-docker"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:       "docker",
		Executable: "my-docker",
		ImageUrl:   "test/example-image:1.0",
		CacheDir:   suite.CacheDir,
		PassEnv:    false,
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	defer func() {
		_, _ = capture.StopCapture()
	}()

	code, err := runtime.Run(types.MoatEnv{Home: suite.HomeDir}, []string{"echo", "hello"}, nil)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, code)

	// The custom binary must have been invoked with the user command.
	contents, err := os.ReadFile(customLogFile)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(contents), "echo\nhello")
}

// TestCacheList tests that CacheList returns the references of all tags
// that docker reports for the repository of the runtime's image URL,
// sorted alphabetically.
func (suite *DockerRuntimeTestSuite) TestCacheList() {
	images, err := suite.newRuntime(false).CacheList()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"test/example-image:1.0", "test/example-image:1.1"}, images)
}

// TestCacheListNoImages tests that CacheList returns an empty list when
// docker reports no images for the repository of the runtime's image URL.
func (suite *DockerRuntimeTestSuite) TestCacheListNoImages() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "none/example-image:1.0",
		CacheDir: suite.CacheDir,
	})

	images, err := runtime.CacheList()

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), images)
}

// TestCacheListUntaggedImage tests that CacheList handles an image URL
// without a tag: the repository is looked up as is.
func (suite *DockerRuntimeTestSuite) TestCacheListUntaggedImage() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "test/example-image",
		CacheDir: suite.CacheDir,
	})

	images, err := runtime.CacheList()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"test/example-image:1.0", "test/example-image:1.1"}, images)
}

// TestCacheListRegistryPort tests that CacheList keeps the port of a
// registry host in the repository lookup.
func (suite *DockerRuntimeTestSuite) TestCacheListRegistryPort() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "localhost:5000/moat:latest",
		CacheDir: suite.CacheDir,
	})

	images, err := runtime.CacheList()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{"localhost:5000/moat:latest"}, images)
}

// TestCacheListEmptyImageUrl tests that CacheList reports an error and
// does not invoke docker when the runtime spec does not set an image URL.
func (suite *DockerRuntimeTestSuite) TestCacheListEmptyImageUrl() {
	runtime := NewDockerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "docker",
		ImageUrl: "",
		CacheDir: suite.CacheDir,
	})

	images, err := runtime.CacheList()

	require.Error(suite.T(), err)
	assert.Nil(suite.T(), images)
	require.NoFileExists(suite.T(), suite.LogFile)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDockerRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(DockerRuntimeTestSuite))
}
