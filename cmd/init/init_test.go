package init_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/AaltoRSE/moat/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InitTestSuite struct {
	suite.Suite
	HomeDir    string
	WorkDir    string
	ConfigPath string
}

func (suite *InitTestSuite) SetupTest() {
	// Point the user home directory at an empty temporary directory so
	// that the init command operates on an isolated configuration.
	suite.HomeDir = suite.T().TempDir()
	suite.T().Setenv("HOME", suite.HomeDir)
	suite.ConfigPath = filepath.Join(suite.HomeDir, ".config", "moat", "moat-config.yaml")

	// Move the working directory to an empty temporary directory so
	// that the "." config search path cannot contain a config file.
	wd, err := os.Getwd()
	require.NoError(suite.T(), err)
	suite.WorkDir = suite.T().TempDir()
	require.NoError(suite.T(), os.Chdir(suite.WorkDir))
	suite.T().Cleanup(func() { _ = os.Chdir(wd) })
}

// runInit executes the moat init command with the given extra arguments
// and returns its standard output.
func (suite *InitTestSuite) runInit(args ...string) string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs(append([]string{"init"}, args...))
	require.NoError(suite.T(), rootCmd.Execute())
	output, err := capture.StopCapture()
	require.NoError(suite.T(), err)
	return output
}

// newTempOutputFile creates an empty temporary file under
// tests.MoatTestDir to be used as the --output path of the init command.
func (suite *InitTestSuite) newTempOutputFile() string {
	f, err := os.CreateTemp(tests.MoatTestDir, "moat-init.*.yaml")
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), f.Close())
	suite.T().Cleanup(func() { assert.NoError(suite.T(), os.Remove(f.Name())) })
	return f.Name()
}

// TestInitCreatesDefaultConfig tests that the init command creates the
// configuration directory and the default configuration file at the
// default location when no configuration exists.
func (suite *InitTestSuite) TestInitCreatesDefaultConfig() {
	output := suite.runInit()

	// The configuration directory and file must have been created.
	assert.DirExists(suite.T(), filepath.Dir(suite.ConfigPath))
	assert.FileExists(suite.T(), suite.ConfigPath)
	assert.Contains(suite.T(), output, suite.ConfigPath)

	// The written configuration must load and hold the default contents.
	cfg, err := config.InitConfig(suite.ConfigPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer", cfg.GetString("defaults.runtime"))
	assert.Equal(suite.T(), "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	// The default image URL is derived from the build-time version.
	imageTag := version.MoatVersion
	if imageTag == "" {
		imageTag = "latest"
	}
	assert.Equal(suite.T(), fmt.Sprintf("ghcr.io/aaltorse/moat:%s", imageTag), cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(suite.T(), "$HOME/.cache/moat/images", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.True(suite.T(), cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.Empty(suite.T(), cfg.GetStringMap("envs"))
}

// TestInitExistingConfig tests that the init command reports the path of
// a pre-existing global configuration and leaves it untouched.
func (suite *InitTestSuite) TestInitExistingConfig() {
	// Create a pre-existing configuration at the default location.
	env := types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	configFile, baseOutput, err := tests.CreateTempConfig(map[string]types.MoatEnv{"test": env}, nil)
	require.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), os.Remove(configFile))
	}()

	require.NoError(suite.T(), os.MkdirAll(filepath.Dir(suite.ConfigPath), 0o755))
	require.NoError(suite.T(), os.WriteFile(suite.ConfigPath, []byte(baseOutput), 0o644))

	output := suite.runInit()

	// The command must report that the configuration exists and give its path.
	assert.Contains(suite.T(), output, "already exists")
	assert.Contains(suite.T(), output, suite.ConfigPath)

	// The pre-existing configuration must be left untouched.
	contents, err := os.ReadFile(suite.ConfigPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), baseOutput, string(contents))
}

// TestInitExistingLocalConfig tests that the init command refuses to
// initialize a configuration when a configuration file is present in the
// current working directory.
func (suite *InitTestSuite) TestInitExistingLocalConfig() {
	// Create a configuration file in the working directory, which is
	// one of the config search locations.
	env := types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	configFile, localOutput, err := tests.CreateTempConfig(map[string]types.MoatEnv{"test": env}, nil)
	require.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), os.Remove(configFile))
	}()

	localConfig := filepath.Join(suite.WorkDir, "moat-config.yaml")
	require.NoError(suite.T(), os.WriteFile(localConfig, []byte(localOutput), 0o644))
	defer func() {
		assert.NoError(suite.T(), os.Remove(localConfig))
	}()

	output := suite.runInit()

	// The command must report the local configuration and must not
	// create the global configuration.
	assert.Contains(suite.T(), output, "already exists")
	assert.Contains(suite.T(), output, localConfig)
	assert.NoFileExists(suite.T(), suite.ConfigPath)
}

// TestInitExistingConfigFlag tests that the init command refuses to
// initialize a configuration when an existing configuration file is given
// via the --config flag.
func (suite *InitTestSuite) TestInitExistingConfigFlag() {
	env := types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	configFile, _, err := tests.CreateTempConfig(map[string]types.MoatEnv{"test": env}, nil)
	require.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), os.Remove(configFile))
	}()

	output := suite.runInit("--config", configFile)

	// The command must report the configuration given by the flag and
	// must not create the global configuration.
	assert.Contains(suite.T(), output, "already exists")
	assert.Contains(suite.T(), output, configFile)
	assert.NoFileExists(suite.T(), suite.ConfigPath)
}

// TestInitOutputFlag tests that the init command writes the default
// configuration to the path given by the --output flag when no
// configuration exists.
func (suite *InitTestSuite) TestInitOutputFlag() {
	outputPath := suite.newTempOutputFile()

	output := suite.runInit("-o", outputPath)

	// The configuration must have been written to the output path.
	assert.FileExists(suite.T(), outputPath)
	assert.Contains(suite.T(), output, outputPath)

	// The written configuration must load and hold the default contents.
	cfg, err := config.InitConfig(outputPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer", cfg.GetString("defaults.runtime"))
	assert.Empty(suite.T(), cfg.GetStringMap("envs"))

	// The global configuration must not have been created.
	assert.NoFileExists(suite.T(), suite.ConfigPath)
}

// TestInitOutputFlagWithFoundConfig tests that the init command switches
// the configuration path to the --output path and writes the found
// configuration there when a configuration file has been found.
func (suite *InitTestSuite) TestInitOutputFlagWithFoundConfig() {
	// Create a configuration file in the working directory, which is
	// one of the config search locations.
	env := types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	configFile, localOutput, err := tests.CreateTempConfig(map[string]types.MoatEnv{"test": env}, nil)
	require.NoError(suite.T(), err)
	defer func() {
		assert.NoError(suite.T(), os.Remove(configFile))
	}()

	localConfig := filepath.Join(suite.WorkDir, "moat-config.yaml")
	require.NoError(suite.T(), os.WriteFile(localConfig, []byte(localOutput), 0o644))
	defer func() {
		assert.NoError(suite.T(), os.Remove(localConfig))
	}()

	outputPath := suite.newTempOutputFile()

	output := suite.runInit("-o", outputPath)

	// The configuration must have been written to the output path.
	assert.FileExists(suite.T(), outputPath)
	assert.Contains(suite.T(), output, outputPath)

	// The written configuration must hold the contents of the found
	// configuration.
	cfg, err := config.InitConfig(outputPath)
	require.NoError(suite.T(), err)
	envs := cfg.GetStringMap("envs")
	assert.Len(suite.T(), envs, 1)
	_, ok := envs["test"]
	assert.True(suite.T(), ok)

	// The found configuration must be left untouched.
	contents, err := os.ReadFile(localConfig)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), localOutput, string(contents))

	// The global configuration must not have been created.
	assert.NoFileExists(suite.T(), suite.ConfigPath)
}

// TestInitOutputFlagOverwritesExistingFile tests that the init command
// overwrites a pre-existing file at the --output path with the
// configuration when no configuration has been found.
func (suite *InitTestSuite) TestInitOutputFlagOverwritesExistingFile() {
	outputPath := suite.newTempOutputFile()
	original := "defaults:\n  runtime: apptainer\n"
	require.NoError(suite.T(), os.WriteFile(outputPath, []byte(original), 0o644))

	output := suite.runInit("-o", outputPath)

	// The configuration must have been written to the output path.
	assert.FileExists(suite.T(), outputPath)
	assert.Contains(suite.T(), output, outputPath)

	// The file must now hold the default configuration contents.
	cfg, err := config.InitConfig(outputPath)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "apptainer", cfg.GetString("defaults.runtime"))
	assert.Empty(suite.T(), cfg.GetStringMap("envs"))

	// The global configuration must not have been created.
	assert.NoFileExists(suite.T(), suite.ConfigPath)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
