package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedDefaultImageURL returns the default apptainer image URL derived
// from the build-time version, mirroring registerDefaults.
func expectedDefaultImageURL() string {
	tag := version.MoatVersion
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("ghcr.io/aaltorse/moat:%s", tag)
}

// TestInitConfigNoConfigFile verifies that InitConfig does not return an
// error when an explicitly given configuration file is not present and
// that the returned configuration holds the default configuration contents.
func TestInitConfigNoConfigFile(t *testing.T) {
	cfgFile := filepath.Join(t.TempDir(), "moat-config.yaml")

	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// The explicitly given path is reported even though no file was loaded.
	assert.Equal(t, cfgFile, cfg.ConfigFileUsed())

	// The default configuration contents must be present.
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtime"))
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	assert.Equal(t, expectedDefaultImageURL(), cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "$HOME/.cache/moat/images", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.Empty(t, cfg.GetStringMap("envs"))

	// The default configuration contents must be retrievable as a string.
	assert.Contains(t, GetConfigAsString(cfg), "apptainer")
}

// TestInitConfigNoConfigFileInSearchPaths verifies that InitConfig does not
// return an error when no configuration file is found in any of the search
// locations and that the returned configuration holds the default
// configuration contents.
func TestInitConfigNoConfigFileInSearchPaths(t *testing.T) {
	// Point the user home directory at an empty temporary directory so
	// that $HOME/.config/moat cannot contain a config file.
	t.Setenv("HOME", t.TempDir())

	// Move the working directory to an empty temporary directory so that
	// the "." search path cannot contain a config file either.
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	cfg, err := InitConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// No configuration file was loaded.
	assert.Empty(t, cfg.ConfigFileUsed())

	// The default configuration contents must be present.
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtime"))
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	assert.Equal(t, expectedDefaultImageURL(), cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "$HOME/.cache/moat/images", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.Empty(t, cfg.GetStringMap("envs"))
}

// TestSetConfigPathSameSourceAndOutput verifies that SetConfigPath
// returns an error when the output path is the same as the path of the
// found configuration and leaves the configuration file path untouched.
func TestSetConfigPathSameSourceAndOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Create and load a configuration file.
	cfgFile := filepath.Join(t.TempDir(), "moat-config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("defaults:\n  runtime: apptainer\n"), 0o644))
	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)

	changed, err := SetConfigPath(cfg, cfgFile)
	assert.Error(t, err)
	assert.False(t, changed)

	// The configuration file path must be left untouched.
	assert.Equal(t, cfgFile, cfg.ConfigFileUsed())
}

// TestSetConfigPathSwitchesToOutput verifies that SetConfigPath sets the
// configuration file path to the output path when a configuration file
// has been found, and that WriteConfig then writes the configuration to
// the output path, creating the parent directory, while leaving the
// source file untouched.
func TestSetConfigPathSwitchesToOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Create and load a configuration file.
	cfgFile := filepath.Join(t.TempDir(), "moat-config.yaml")
	original := "defaults:\n  runtime: apptainer\n"
	require.NoError(t, os.WriteFile(cfgFile, []byte(original), 0o644))
	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)

	outputFile := filepath.Join(t.TempDir(), "nested", "moat-config.yaml")
	changed, err := SetConfigPath(cfg, outputFile)
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, outputFile, cfg.ConfigFileUsed())

	// The configuration must be written to the output path.
	require.NoError(t, WriteConfig(cfg))
	assert.FileExists(t, outputFile)

	// The source configuration file must be left untouched.
	contents, err := os.ReadFile(cfgFile)
	require.NoError(t, err)
	assert.Equal(t, original, string(contents))
}
