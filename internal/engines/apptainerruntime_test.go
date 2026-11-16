package engines

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
if [ "$1" = "sif" ] && [ "$2" = "list" ]; then
    printf '%s\n' '2 | 1 | NONE | 32768-65536 | FS (Squashfs/Data/amd64)' '5 | 1 | NONE | 65536-98304 | FS (Squashfs/*System/amd64)'
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
// run --contain --pid --no-home --no-mount cwd, binds the fake home
// directory, and ends with the image path and the user command.
func (suite *ApptainerRuntimeTestSuite) TestRunBaseArgs() {
	home := filepath.Join(suite.HomeDir, "fake-home")
	require.NoError(suite.T(), os.MkdirAll(home, 0o755))
	env := types.MoatEnv{Home: home}

	apptainerArgs := suite.runAndCaptureArgs(suite.newRuntime(false), env, "echo", "hello")

	// The base apptainer arguments must be present.
	assert.Equal(suite.T(), "run", apptainerArgs[0])
	assert.Contains(suite.T(), apptainerArgs, "--contain")
	assert.Contains(suite.T(), apptainerArgs, "--pid")
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

// TestNewRuntimeDefaultsExecutable tests that the runtime defaults to the
// apptainer executable when the runtime spec does not set one.
func (suite *ApptainerRuntimeTestSuite) TestNewRuntimeDefaultsExecutable() {
	runtime := NewApptainerRuntimeFromSpec(&types.RuntimeSpec{Type: "apptainer"})

	assert.Equal(suite.T(), "apptainer", runtime.Executable)
}

// TestNewRuntimeUsesSpecExecutable tests that the runtime uses the
// executable given in the runtime spec.
func (suite *ApptainerRuntimeTestSuite) TestNewRuntimeUsesSpecExecutable() {
	runtime := NewApptainerRuntimeFromSpec(&types.RuntimeSpec{Type: "apptainer", Executable: "my-apptainer"})

	assert.Equal(suite.T(), "my-apptainer", runtime.Executable)
}

// TestRunUsesSpecExecutable tests that Run invokes the executable given
// in the runtime spec instead of the default apptainer binary.
func (suite *ApptainerRuntimeTestSuite) TestRunUsesSpecExecutable() {
	// Install a fake binary under a custom name on the PATH.
	fakeBinDir := suite.T().TempDir()
	customLogFile := filepath.Join(fakeBinDir, "custom-args.log")
	script := strings.ReplaceAll(fakeApptainerScript, "<logFile>", customLogFile)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "my-apptainer"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	runtime := NewApptainerRuntimeFromSpec(&types.RuntimeSpec{
		Type:       "apptainer",
		Executable: "my-apptainer",
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

// TestCacheList tests that CacheList returns the paths of the .sif files
// in the cache directory, sorted alphabetically, and ignores other files
// and directories.
func (suite *ApptainerRuntimeTestSuite) TestCacheList() {
	require.NoError(suite.T(), os.WriteFile(filepath.Join(suite.CacheDir, "second.sif"), nil, 0o644))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(suite.CacheDir, "first.sif"), nil, 0o644))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(suite.CacheDir, "notes.txt"), nil, 0o644))
	require.NoError(suite.T(), os.MkdirAll(filepath.Join(suite.CacheDir, "subdir.sif"), 0o755))

	images, err := suite.newRuntime(false).CacheList()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), []string{
		filepath.Join(suite.CacheDir, "first.sif"),
		filepath.Join(suite.CacheDir, "second.sif"),
	}, images)
}

// TestCacheListEmptyCacheDir tests that CacheList returns an empty list
// when the cache directory exists but contains no images.
func (suite *ApptainerRuntimeTestSuite) TestCacheListEmptyCacheDir() {
	images, err := suite.newRuntime(false).CacheList()

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), images)
}

// TestCacheListMissingCacheDir tests that CacheList returns an empty list
// when the cache directory does not exist.
func (suite *ApptainerRuntimeTestSuite) TestCacheListMissingCacheDir() {
	runtime := NewApptainerRuntimeFromSpec(&types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "test/example-image:1.0",
		CacheDir: filepath.Join(suite.CacheDir, "missing"),
	})

	images, err := runtime.CacheList()

	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), images)
}

// newOverlayRuntime builds an ApptainerRuntime with the given overlay
// owner, driver, and overlay directory settings.
func (suite *ApptainerRuntimeTestSuite) newOverlayRuntime(owner, driver, overlayDir string) *ApptainerRuntime {
	runtime := suite.newRuntime(false)
	runtime.OverlayOwner = owner
	runtime.OverlayDriver = driver
	runtime.OverlayDir = overlayDir
	return runtime
}

// overlayFlagValues returns the argument following each occurrence of flag.
func overlayFlagValues(args []string, flag string) []string {
	var values []string
	for i := 0; i+1 < len(args); i++ {
		if args[i] == flag {
			values = append(values, args[i+1])
		}
	}
	return values
}

// TestRunNoOverlayArgsWithoutOverlayMounts verifies that normal runs keep
// the original invocation when the runtime configures overlay storage.
func (suite *ApptainerRuntimeTestSuite) TestRunNoOverlayArgsWithoutOverlayMounts() {
	root := filepath.Join(suite.T().TempDir(), "overlays")
	runtime := suite.newOverlayRuntime("external", "kernel-overlayfs", root)
	args := suite.runAndCaptureArgs(runtime, types.MoatEnv{Home: suite.HomeDir}, "echo", "hello")
	assert.NotContains(suite.T(), args, "--no-eval")
	assert.NotContains(suite.T(), args, "--mount")
	assert.NotContains(suite.T(), args, "--fusemount")
	assert.Equal(suite.T(), []string{"echo", "hello"}, args[len(args)-2:])
	assert.NoDirExists(suite.T(), root)
}

// TestRunInternalOwnerUsesFusemount verifies that the default driver reads
// the SIF root from a separate mount and supplies temporary writable layers.
func (suite *ApptainerRuntimeTestSuite) TestRunInternalOwnerUsesFusemount() {
	work := suite.T().TempDir()
	suite.T().Setenv("TMPDIR", work)
	runtime := suite.newOverlayRuntime("", "", "")
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{"/opt/conda", "/usr/local"}}

	args := suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	assert.Equal(suite.T(), "exec", args[0])
	fusemounts := overlayFlagValues(args, "--fusemount")
	require.Len(suite.T(), fusemounts, 2)
	match := regexp.MustCompile(`lowerdir=(/moat-overlay-[0-9]+)/image/opt/conda `).FindStringSubmatch(fusemounts[0])
	require.Len(suite.T(), match, 2)
	stage := match[1]
	for i, dest := range env.OverlayMounts {
		assert.Equal(suite.T(), fmt.Sprintf("container:fuse-overlayfs -o lowerdir=%s/image%s -o upperdir=%s/%d/upper -o workdir=%s/%d/work %s",
			stage, dest, stage, i, stage, i, dest), fusemounts[i])
	}
	image, err := runtime.GetImage()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), overlayFlagValues(args, "--mount"), mountSpec(image.Path, stage+"/image", "ro", "image-src=/", "id=5"))
	scratch := filepath.Join(work, filepath.Base(stage))
	assert.Contains(suite.T(), overlayFlagValues(args, "--mount"), mountSpec(filepath.Join(scratch, "0"), stage+"/0"))
	assert.NoDirExists(suite.T(), scratch)
	assert.NotContains(suite.T(), args, "--overlay")
	assert.Equal(suite.T(), []string{"/bin/sh", stage + "/run.sh", "fuse-overlayfs", "2",
		stage + "/0", "/opt/conda", stage + "/1", "/usr/local", "echo", "hello"}, args[len(args)-10:])
}

// TestRunInternalOwnerWithSourceUsesPersistentUpper verifies that the
// driver receives a container path to the explicitly bound host storage.
func (suite *ApptainerRuntimeTestSuite) TestRunInternalOwnerWithSourceUsesPersistentUpper() {
	runtime := suite.newOverlayRuntime("internal", "fuse-overlayfs", "")
	source := filepath.Join(suite.T().TempDir(), "my overlay, storage")
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{source + ":/opt/conda"}}

	args := suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	fusemounts := overlayFlagValues(args, "--fusemount")
	require.Len(suite.T(), fusemounts, 1)
	match := regexp.MustCompile(`upperdir=(/moat-overlay-[0-9]+/0)/upper `).FindStringSubmatch(fusemounts[0])
	require.Len(suite.T(), match, 2)
	assert.Contains(suite.T(), overlayFlagValues(args, "--mount"), mountSpec(source, match[1]))
	assert.NotContains(suite.T(), fusemounts[0], source)
	assert.DirExists(suite.T(), filepath.Join(source, "upper"))
	assert.DirExists(suite.T(), filepath.Join(source, "work"))
}

// TestRunExternalKernelOverlayUsesDistinctStorage verifies that destinations
// with the same old storage name get independent kernel overlays.
func (suite *ApptainerRuntimeTestSuite) TestRunExternalKernelOverlayUsesDistinctStorage() {
	root := suite.T().TempDir()
	runtime := suite.newOverlayRuntime("external", "kernel-overlayfs", root)
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{"/opt/conda", "/opt_conda"}}

	args := suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	assert.NotContains(suite.T(), args, "--overlay")
	assert.NotContains(suite.T(), args, "--fusemount")
	assert.Contains(suite.T(), args, "/usr/bin/unshare")
	stage := regexp.MustCompile(`^/moat-overlay-[0-9]+/0$`).FindString(args[len(args)-6])
	require.NotEmpty(suite.T(), stage)
	assert.Equal(suite.T(), []string{stage, "/opt/conda", filepath.Dir(stage) + "/1", "/opt_conda", "echo", "hello"}, args[len(args)-6:])
	dirs, err := os.ReadDir(filepath.Join(root, "testenv"))
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dirs, 2)
	for _, dir := range dirs {
		assert.DirExists(suite.T(), filepath.Join(root, "testenv", dir.Name(), "upper"))
		assert.DirExists(suite.T(), filepath.Join(root, "testenv", dir.Name(), "work"))
	}
}

// TestRunExternalFuseOverlayUsesPersistentStorage verifies that an external
// FUSE mount with no explicit source uses the managed storage directory.
func (suite *ApptainerRuntimeTestSuite) TestRunExternalFuseOverlayUsesPersistentStorage() {
	root := suite.T().TempDir()
	runtime := suite.newOverlayRuntime("external", "", root)
	env := types.MoatEnv{Home: suite.HomeDir, OverlayMounts: []string{"/opt/conda"}}
	args := suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	assert.Len(suite.T(), overlayFlagValues(args, "--fusemount"), 1)
	dirs, err := os.ReadDir(filepath.Join(root, "default"))
	require.NoError(suite.T(), err)
	require.Len(suite.T(), dirs, 1)
	upper := filepath.Join(root, "default", dirs[0].Name(), "upper")
	data := filepath.Join(upper, "keep")
	require.NoError(suite.T(), os.WriteFile(data, []byte("persist"), 0o600))
	suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	assert.FileExists(suite.T(), data)
}

// TestRunRejectsSharedOverlayStorage verifies that overlapping runs cannot
// reuse an active writable layer, and that closing the first plan frees it.
func (suite *ApptainerRuntimeTestSuite) TestRunRejectsSharedOverlayStorage() {
	runtime := suite.newOverlayRuntime("external", "kernel-overlayfs", "")
	source := suite.T().TempDir()
	env := types.MoatEnv{Home: suite.HomeDir, OverlayMounts: []string{source + ":/data"}}
	plan, err := runtime.prepareOverlays(env, "", false)
	require.NoError(suite.T(), err)
	_, err = runtime.prepareOverlays(env, "", false)
	require.ErrorContains(suite.T(), err, "exclusive storage")
	plan.close()
	suite.runAndCaptureArgs(runtime, env, "echo", "hello")
}

// fakeBtrfsScript is the body of a fake btrfs executable used by the btrfs
// overlay tests. "subvolume create <dir>" marks the directory as a
// subvolume. The read-only property query succeeds for marked directories.
// Every invocation is appended to the log file whose path is substituted
// for <logFile>.
const fakeBtrfsScript = `#!/bin/sh
printf '%s\n' "$@" >> "<logFile>"
if [ "$1" = "subvolume" ] && [ "$2" = "create" ]; then
    mkdir -p "$3" && : > "$3/.subvolume"
    exit 0
fi
if [ "$1" = "property" ] && [ "$2" = "get" ]; then
    if [ -f "$5/.subvolume" ]; then
        printf 'ro=false\n'
        exit 0
    fi
    exit 1
fi
exit 0
`

// TestRunExternalBtrfsOverlayCreatesSubvolume verifies that explicit btrfs
// storage is a subvolume with kernel overlay backing directories.
func (suite *ApptainerRuntimeTestSuite) TestRunExternalBtrfsOverlayCreatesSubvolume() {
	// Install a fake btrfs binary on the PATH.
	fakeBinDir := suite.T().TempDir()
	btrfsLog := filepath.Join(fakeBinDir, "btrfs-args.log")
	script := strings.ReplaceAll(fakeBtrfsScript, "<logFile>", btrfsLog)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "btrfs"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	source := filepath.Join(suite.T().TempDir(), "source")
	runtime := suite.newOverlayRuntime("external", "btrfs", "")
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{source + ":/opt/conda"}}
	args := suite.runAndCaptureArgs(runtime, env, "echo", "hello")
	assert.Contains(suite.T(), args, "/usr/bin/unshare")
	assert.NotContains(suite.T(), args, "--overlay")
	assert.DirExists(suite.T(), filepath.Join(source, "upper"))
	assert.DirExists(suite.T(), filepath.Join(source, "work"))
	contents, err := os.ReadFile(btrfsLog)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), string(contents), "subvolume\ncreate\n"+source)
}

// TestRunExternalBtrfsOverlayExistingNonSubvolumeFails tests that the
// btrfs driver fails with an error when the overlay directory exists but
// is not a btrfs subvolume.
func (suite *ApptainerRuntimeTestSuite) TestRunExternalBtrfsOverlayExistingNonSubvolumeFails() {
	fakeBinDir := suite.T().TempDir()
	btrfsLog := filepath.Join(fakeBinDir, "btrfs-args.log")
	script := strings.ReplaceAll(fakeBtrfsScript, "<logFile>", btrfsLog)
	require.NoError(suite.T(), os.WriteFile(filepath.Join(fakeBinDir, "btrfs"), []byte(script), 0o755))
	suite.T().Setenv("PATH", fakeBinDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	source := suite.T().TempDir() // exists, but is not a subvolume
	runtime := suite.newOverlayRuntime("external", "btrfs", "")
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{source + ":/opt/conda"}}

	code, err := runtime.Run(env, []string{"echo", "hello"}, nil)
	require.Error(suite.T(), err)
	assert.Equal(suite.T(), 1, code)
	assert.Contains(suite.T(), err.Error(), "is a btrfs subvolume")
}

// TestRunInternalOwnerRejectsKernelDriver tests that the internal owner
// combined with a non-fuse driver fails with an error instead of running.
func (suite *ApptainerRuntimeTestSuite) TestRunInternalOwnerRejectsKernelDriver() {
	runtime := suite.newOverlayRuntime("internal", "kernel-overlayfs", "")
	env := types.MoatEnv{Home: suite.HomeDir, Name: "testenv", OverlayMounts: []string{"/opt/conda"}}

	code, err := runtime.Run(env, []string{"echo", "hello"}, nil)
	require.Error(suite.T(), err)
	assert.Equal(suite.T(), 1, code)
	assert.Contains(suite.T(), err.Error(), "only supports overlaydriver")
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestApptainerRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(ApptainerRuntimeTestSuite))
}
