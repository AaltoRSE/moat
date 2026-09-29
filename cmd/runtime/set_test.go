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

// createSetRuntime creates a user runtime named "custom" with the runtime
// create command and returns its specification, so that the set tests can
// start from a configuration containing a user runtime.
func (suite *RuntimeTestSuite) createSetRuntime() types.RuntimeSpec {
	spec := types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:custom",
		CacheDir: tests.MoatTestDir + "/runtime_set_cache",
		PassEnv:  true,
		MountCWD: false,
	}

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create",
		"--name", "custom", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	return spec
}

// TestSetImageurl tests that runtime set updates the imageurl variable of
// an existing user runtime and leaves the other variables of the runtime
// untouched.
func (suite *RuntimeTestSuite) TestSetImageurl() {
	spec := suite.createSetRuntime()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--imageurl", "ghcr.io/aaltorse/moat:set"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set custom.imageurl = [ghcr.io/aaltorse/moat:set]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")
	assert.NotContains(suite.T(), capturedOutput, "Overriding default runtime")

	// The updated runtime must have the new image URL and keep the other
	// variables of the original specification
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), "ghcr.io/aaltorse/moat:set", updatedSpec.ImageUrl)
	assert.Equal(suite.T(), spec.Type, updatedSpec.Type)
	assert.Equal(suite.T(), spec.CacheDir, updatedSpec.CacheDir)
	assert.Equal(suite.T(), spec.PassEnv, updatedSpec.PassEnv)
	assert.Equal(suite.T(), spec.MountCWD, updatedSpec.MountCWD)
}

// TestSetCachedir tests that runtime set with a relative --cachedir path
// stores the cache directory as an absolute path.
func (suite *RuntimeTestSuite) TestSetCachedir() {
	suite.createSetRuntime()
	cacheDir := "runtime_set_new_cache"

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--cachedir", cacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// The stored cache directory must be resolved to an absolute path
	absCacheDir, err := filepath.Abs(cacheDir)
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set custom.cachedir = ["+absCacheDir+"]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")

	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), absCacheDir, updatedSpec.CacheDir)
}

// TestSetPassEnv tests that runtime set with the --passenv flag stores the
// boolean setting of the runtime.
func (suite *RuntimeTestSuite) TestSetPassEnv() {
	suite.createSetRuntime()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--passenv"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set custom.passenv = [true]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")

	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.True(suite.T(), updatedSpec.PassEnv)
}

// TestSetPassEnvFalse tests that runtime set with --passenv=false fails
// validation and does not modify the configuration.
func (suite *RuntimeTestSuite) TestSetPassEnvFalse() {
	spec := suite.createSetRuntime()
	contentsAfterCreate, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--passenv=false"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.NotContains(suite.T(), capturedOutput, "Runtime updated successfully.")

	// The original runtime spec and the config file must be left unchanged
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec, updatedSpec)
	contents, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), string(contentsAfterCreate), string(contents))
}

// TestSetMountCWD tests that runtime set with the --mountcwd flag stores
// the boolean setting of the runtime.
func (suite *RuntimeTestSuite) TestSetMountCWD() {
	suite.createSetRuntime()

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--mountcwd"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set custom.mountcwd = [true]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")

	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.True(suite.T(), updatedSpec.MountCWD)
	assert.True(suite.T(), updatedSpec.PassEnv)
}

// TestSetMultipleFlags tests that runtime set applies several given flags
// in a single invocation and leaves the other variables of the runtime
// untouched.
func (suite *RuntimeTestSuite) TestSetMultipleFlags() {
	suite.createSetRuntime()
	newCacheDir := filepath.Join(tests.MoatTestDir, "runtime_set_multi_cache")

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom",
		"--type", "apptainer", "--imageurl", "ghcr.io/aaltorse/moat:set", "--cachedir", newCacheDir, "--mountcwd"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set custom.type = [apptainer]")
	assert.Contains(suite.T(), capturedOutput, "Set custom.imageurl = [ghcr.io/aaltorse/moat:set]")
	assert.Contains(suite.T(), capturedOutput, "Set custom.cachedir = ["+newCacheDir+"]")
	assert.Contains(suite.T(), capturedOutput, "Set custom.mountcwd = [true]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")

	// The updated runtime must have all the given values and keep the
	// variables that were not given
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:set",
		CacheDir: newCacheDir,
		PassEnv:  true,
		MountCWD: true,
	}, updatedSpec)
}

// TestSetDefaultRuntime tests that runtime set with the name of a default
// runtime creates a user runtime seeded from the default runtime's
// specification with the given change applied, and that the user runtime
// overwrites the default runtime in runtime lookups.
func (suite *RuntimeTestSuite) TestSetDefaultRuntime() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "apptainer", "--imageurl", "ghcr.io/aaltorse/moat:set"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Overriding default runtime 'apptainer'")
	assert.Contains(suite.T(), capturedOutput, "Set apptainer.imageurl = [ghcr.io/aaltorse/moat:set]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")

	// A user runtime must have been created with the default runtime's
	// specification, the given change applied, and the default cache
	// directory resolved to an absolute path
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	userSpec, exists := config.GetUserRuntimes(cfg)["apptainer"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), "apptainer", userSpec.Type)
	assert.Equal(suite.T(), "ghcr.io/aaltorse/moat:set", userSpec.ImageUrl)
	home, err := os.UserHomeDir()
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), filepath.Join(home, ".cache", "moat", "images"), userSpec.CacheDir)
	assert.True(suite.T(), userSpec.PassEnv)
	assert.False(suite.T(), userSpec.MountCWD)

	// The user runtime must overwrite the default runtime in lookups
	allRuntimes := config.GetRuntimes(cfg)
	assert.Len(suite.T(), allRuntimes, 1)
	assert.Equal(suite.T(), "ghcr.io/aaltorse/moat:set", allRuntimes["apptainer"].ImageUrl)
}

// TestSetUserRuntimeWithDefaultName tests that runtime set with the name
// of a user runtime that already overwrites a default runtime updates the
// user runtime in place without creating a new one.
func (suite *RuntimeTestSuite) TestSetUserRuntimeWithDefaultName() {
	spec := types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:custom",
		CacheDir: tests.MoatTestDir + "/runtime_set_apptainer_cache",
		PassEnv:  true,
		MountCWD: false,
	}

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create",
		"--name", "apptainer", "--type", spec.Type, "--imageurl", spec.ImageUrl, "--cachedir", spec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "apptainer", "--mountcwd"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.Contains(suite.T(), capturedOutput, "Set apptainer.mountcwd = [true]")
	assert.Contains(suite.T(), capturedOutput, "Runtime updated successfully.")
	assert.NotContains(suite.T(), capturedOutput, "Overriding default runtime")

	// The existing user runtime must be updated in place
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	updatedSpec, exists := config.GetUserRuntimes(cfg)["apptainer"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), spec.ImageUrl, updatedSpec.ImageUrl)
	assert.Equal(suite.T(), spec.CacheDir, updatedSpec.CacheDir)
	assert.True(suite.T(), updatedSpec.MountCWD)
}

// TestSetNonexistentRuntime tests that runtime set with the name of a
// runtime that does not exist does not modify the configuration.
func (suite *RuntimeTestSuite) TestSetNonexistentRuntime() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "nosuch", "--imageurl", "ghcr.io/aaltorse/moat:set"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.NotContains(suite.T(), capturedOutput, "Runtime updated successfully.")

	// No runtime must have been created and the config file must be
	// unmodified
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), config.GetUserRuntimes(cfg))
	contents, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.BaseOutput, string(contents))
}

// TestSetNoFlags tests that runtime set without any variable flag does not
// modify the configuration.
func (suite *RuntimeTestSuite) TestSetNoFlags() {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "apptainer"})
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	assert.NotContains(suite.T(), capturedOutput, "Runtime updated successfully.")

	// No runtime must have been created and the config file must be
	// unmodified
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Empty(suite.T(), config.GetUserRuntimes(cfg))
	contents, err := os.ReadFile(suite.ConfigFile)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.BaseOutput, string(contents))
}

// TestSetLeavesOtherRuntimesUnchanged tests that runtime set only modifies
// the runtime given by the --name flag and leaves other runtimes untouched.
func (suite *RuntimeTestSuite) TestSetLeavesOtherRuntimesUnchanged() {
	spec := suite.createSetRuntime()

	// Create a second user runtime to verify it stays untouched
	otherSpec := types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:other",
		CacheDir: tests.MoatTestDir + "/runtime_set_other_cache",
		PassEnv:  true,
		MountCWD: false,
	}
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "create",
		"--name", "other", "--type", otherSpec.Type, "--imageurl", otherSpec.ImageUrl, "--cachedir", otherSpec.CacheDir})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err := capture.StopCapture()
	require.NoError(suite.T(), err)

	// Set a variable on the "custom" runtime
	capture = utils.OutputCapture{}
	capture.StartCapture()
	rootCmd = root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "runtime", "set", "--name", "custom", "--imageurl", "ghcr.io/aaltorse/moat:set"})
	require.NoError(suite.T(), rootCmd.Execute())
	_, err = capture.StopCapture()
	require.NoError(suite.T(), err)

	// The "custom" runtime must be updated and the "other" runtime must
	// be left untouched
	cfg, err := config.InitConfig(suite.ConfigFile)
	require.NoError(suite.T(), err)
	userRuntimes := config.GetUserRuntimes(cfg)
	updatedSpec, exists := userRuntimes["custom"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), "ghcr.io/aaltorse/moat:set", updatedSpec.ImageUrl)
	assert.Equal(suite.T(), spec.Type, updatedSpec.Type)
	assert.Equal(suite.T(), spec.CacheDir, updatedSpec.CacheDir)
	otherUpdatedSpec, exists := userRuntimes["other"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), otherSpec, otherUpdatedSpec)
}
