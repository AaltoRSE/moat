package cache_test

import (
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cachedRuntimeSpec returns a valid apptainer runtime specification whose
// cache directory is the given path.
func cachedRuntimeSpec(cacheDir string) types.RuntimeSpec {
	return types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:custom",
		CacheDir: cacheDir,
		PassEnv:  true,
	}
}

// createCacheDir creates a cache directory under tests.MoatTestDir
// containing the given image file names, and returns its path. The
// directory is removed when the test finishes.
func (suite *CacheTestSuite) createCacheDir(images ...string) string {
	cacheDir := filepath.Join(tests.MoatTestDir, "cachelist_test_cache")
	require.NoError(suite.T(), os.MkdirAll(cacheDir, 0o755))
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.RemoveAll(cacheDir)) })
	for _, image := range images {
		require.NoError(suite.T(), os.WriteFile(filepath.Join(cacheDir, image), nil, 0o644))
	}
	return cacheDir
}

// runCacheList runs `moat runtime cache list` with the given extra
// arguments against the given config file and returns the standard output.
func (suite *CacheTestSuite) runCacheList(configFile string, args ...string) string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs(append([]string{"--config", configFile, "runtime", "cache", "list"}, args...))
	require.NoError(suite.T(), rootCmd.Execute())
	capturedOutput, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	return capturedOutput
}

// TestCacheList tests that runtime cache list prints the images cached in
// the named runtime's cache directory, sorted alphabetically.
func (suite *CacheTestSuite) TestCacheList() {
	cacheDir := suite.createCacheDir("second.sif", "first.sif")
	configFile, _, err := tests.CreateTempConfig(nil, map[string]types.RuntimeSpec{"custom": cachedRuntimeSpec(cacheDir)})
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.Remove(configFile)) })

	output := suite.runCacheList(configFile, "--name", "custom")

	assert.Equal(suite.T(),
		filepath.Join(cacheDir, "first.sif")+"\n"+filepath.Join(cacheDir, "second.sif")+"\n",
		output)
}

// TestCacheListEmpty tests that runtime cache list reports an empty cache
// directory of the named runtime.
func (suite *CacheTestSuite) TestCacheListEmpty() {
	cacheDir := suite.createCacheDir()
	configFile, _, err := tests.CreateTempConfig(nil, map[string]types.RuntimeSpec{"custom": cachedRuntimeSpec(cacheDir)})
	require.NoError(suite.T(), err)
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.Remove(configFile)) })

	output := suite.runCacheList(configFile, "--name", "custom")

	assert.Equal(suite.T(), "No cached images for runtime 'custom'.\n", output)
}

// TestCacheListDefaultRuntime tests that runtime cache list without the
// -n/--name flag lists the images cached by the default runtime.
func (suite *CacheTestSuite) TestCacheListDefaultRuntime() {
	home := filepath.Join(tests.MoatTestDir, "cachelist_test_home")
	cacheDir := filepath.Join(home, ".cache", "moat", "images")
	require.NoError(suite.T(), os.MkdirAll(cacheDir, 0o755))
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.RemoveAll(home)) })
	require.NoError(suite.T(), os.WriteFile(filepath.Join(cacheDir, "cached.sif"), nil, 0o644))
	suite.T().Setenv("HOME", home)

	output := suite.runCacheList(suite.ConfigFile)

	assert.Equal(suite.T(), filepath.Join(cacheDir, "cached.sif")+"\n", output)
}

// TestCacheListWithMissingName tests that runtime cache list with the
// -n/--name flag given a name of a nonexistent runtime prints nothing.
func (suite *CacheTestSuite) TestCacheListWithMissingName() {
	output := suite.runCacheList(suite.ConfigFile, "--name", "nonexistent")

	assert.Empty(suite.T(), output)
}
