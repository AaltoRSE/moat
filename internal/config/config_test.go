package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	assert.Equal(t, "ghcr.io/aaltorse/vscode-apptainer:latest", cfg.GetString("defaults.runtimes.apptainer.imageurl"))
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
	assert.Equal(t, "ghcr.io/aaltorse/vscode-apptainer:latest", cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "$HOME/.cache/moat/images", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.Empty(t, cfg.GetStringMap("envs"))
}
