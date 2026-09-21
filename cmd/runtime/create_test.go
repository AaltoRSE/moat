package runtime_test

import (
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createRuntimeSpec returns the runtime specification used by the create
// tests.
func createRuntimeSpec() types.RuntimeSpec {
	return types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:custom",
		CacheDir: tests.MoatTestDir + "/runtime_create_cache",
		PassEnv:  true,
		MountCWD: false,
	}
}

// TestCreate tests that runtime create with the --type, --imageurl and
// --cachedir flags stores a new runtime in the top-level runtimes key of
// the written config file.
func (suite *RuntimeTestSuite) TestCreate() {
	spec := createRuntimeSpec()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Runtime created successfully.")

	// The new runtime must be stored in the top-level runtimes key of the
	// written config file
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	runtimes := config.GetUserRuntimes(cfg)
	createdSpec, exists := runtimes["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec, createdSpec)

	// The default runtimes must remain available alongside the new runtime
	allRuntimes := config.GetRuntimes(cfg)
	assert.Contains(suite.T(), allRuntimes, "apptainer")
	assert.Contains(suite.T(), allRuntimes, "custom")
}

// TestCreateWithPassEnvMountCWD tests that runtime create with the
// --passenv and --mountcwd flags stores the boolean settings of the
// runtime.
func (suite *RuntimeTestSuite) TestCreateWithPassEnvMountCWD() {
	spec := createRuntimeSpec()
	spec.MountCWD = true

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir, "--passenv", "--mountcwd"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Runtime created successfully.")

	// The stored runtime must have the boolean settings as given
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	createdSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec, createdSpec)
	assert.True(suite.T(), createdSpec.PassEnv)
	assert.True(suite.T(), createdSpec.MountCWD)
}

// TestCreateExisting tests that runtime create with the name of an
// existing user-specified runtime does not modify the configuration.
func (suite *RuntimeTestSuite) TestCreateExisting() {
	spec := createRuntimeSpec()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Try to create the same runtime again with a different image URL
	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", spec.Type, "--imageurl", "ghcr.io/aaltorse/moat:other", "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Runtime already exists.")

	// The original runtime spec must be left unchanged
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	createdSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec, createdSpec)
}

// TestCreateInvalidName tests that runtime create with a name containing
// invalid characters does not create a runtime.
func (suite *RuntimeTestSuite) TestCreateInvalidName() {
	spec := createRuntimeSpec()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "bad-name", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// No runtime must have been created and the config file must be
	// unmodified
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), config.GetUserRuntimes(cfg))
	contents, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.BaseOutput, string(contents))
}

// TestCreateOverwriteDefaultRuntime tests that runtime create with the
// name of a default runtime is allowed and that the new runtime overwrites
// the default runtime in runtime lookups.
func (suite *RuntimeTestSuite) TestCreateOverwriteDefaultRuntime() {
	spec := createRuntimeSpec()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "apptainer", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Runtime created successfully.")

	// The user-specified runtime must be stored and take precedence over
	// the default runtime with the same name
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	createdSpec, exists := config.GetUserRuntimes(cfg)["apptainer"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec, createdSpec)
	allRuntimes := config.GetRuntimes(cfg)
	assert.Len(suite.T(), allRuntimes, 1)
	assert.Equal(suite.T(), spec.ImageUrl, allRuntimes["apptainer"].ImageUrl)
}

// TestCreateRelativeCacheDir tests that runtime create with a relative
// --cachedir path stores the cache directory as an absolute path.
func (suite *RuntimeTestSuite) TestCreateRelativeCacheDir() {
	cacheDir := "runtime_create_rel_cache"

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", "apptainer", "--imageurl", "ghcr.io/aaltorse/moat:custom", "--cachedir", cacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Runtime created successfully.")

	// The stored cache directory must be resolved to an absolute path
	absCacheDir, err := filepath.Abs(cacheDir)
	require.NoError(suite.T(), err)
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	createdSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), absCacheDir, createdSpec.CacheDir)
}

// TestCreateMissingImageurl tests that runtime create with the apptainer
// type but without the --imageurl flag fails configuration validation and
// does not write the new runtime to the config file.
func (suite *RuntimeTestSuite) TestCreateMissingImageurl() {
	spec := createRuntimeSpec()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create", "--name", "custom", "--type", spec.Type, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// No runtime must have been created and the config file must be
	// unmodified
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), config.GetUserRuntimes(cfg))
	contents, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.BaseOutput, string(contents))
}
