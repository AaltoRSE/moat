package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AaltoRSE/moat/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedDefaultImageURL returns the default apptainer image URL derived
// from the build-time version, mirroring CreateDefaultConfig.
func expectedDefaultImageURL() string {
	tag := version.MoatVersion
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("ghcr.io/aaltorse/moat:%s", tag)
}

// clearMoatDefaultEnvVars unsets all MOAT_DEFAULTS_* environment variables
// for the duration of the test and restores them on cleanup, so that the
// default-configuration assertions are independent of the ambient
// environment.
func clearMoatDefaultEnvVars(t *testing.T) {
	t.Helper()
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if !strings.HasPrefix(key, "MOAT_DEFAULTS_") {
			continue
		}
		old, ok := os.LookupEnv(key)
		require.NoError(t, os.Unsetenv(key))
		t.Cleanup(func() {
			if ok {
				_ = os.Setenv(key, old)
			}
		})
	}
}

// TestInitConfigNoConfigFile verifies that InitConfig does not return an
// error when an explicitly given configuration file is not present and
// that the returned configuration holds the default configuration contents.
func TestInitConfigNoConfigFile(t *testing.T) {
	clearMoatDefaultEnvVars(t)
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
	clearMoatDefaultEnvVars(t)
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

// TestInitConfigMoatConfigEnv verifies that InitConfig loads the
// configuration file pointed to by the MOAT_CONFIG environment variable
// when no configuration file path is given.
func TestInitConfigMoatConfigEnv(t *testing.T) {
	// Point the user home directory and working directory at empty
	// temporary directories so that the search paths cannot contain a
	// config file.
	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Create a configuration file and point MOAT_CONFIG at it.
	cfgFile := filepath.Join(t.TempDir(), "moat-config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("defaults:\n  runtime: custom\n"), 0o644))
	t.Setenv("MOAT_CONFIG", cfgFile)

	cfg, err := InitConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// The configuration file from MOAT_CONFIG was loaded.
	assert.Equal(t, cfgFile, cfg.ConfigFileUsed())
	assert.Equal(t, "custom", cfg.GetString("defaults.runtime"))
}

// TestInitConfigExplicitFileOverridesMoatConfigEnv verifies that an
// explicitly given configuration file path takes priority over the
// MOAT_CONFIG environment variable.
func TestInitConfigExplicitFileOverridesMoatConfigEnv(t *testing.T) {
	// Point the user home directory and working directory at empty
	// temporary directories so that the search paths cannot contain a
	// config file.
	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// Create two configuration files and point MOAT_CONFIG at one of
	// them.
	envFile := filepath.Join(t.TempDir(), "env-config.yaml")
	require.NoError(t, os.WriteFile(envFile, []byte("defaults:\n  runtime: fromenv\n"), 0o644))
	t.Setenv("MOAT_CONFIG", envFile)

	explicitFile := filepath.Join(t.TempDir(), "explicit-config.yaml")
	require.NoError(t, os.WriteFile(explicitFile, []byte("defaults:\n  runtime: fromflag\n"), 0o644))

	cfg, err := InitConfig(explicitFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// The explicitly given file was loaded, not the one from MOAT_CONFIG.
	assert.Equal(t, explicitFile, cfg.ConfigFileUsed())
	assert.Equal(t, "fromflag", cfg.GetString("defaults.runtime"))
}

// TestCreateDefaultConfigApptainerRuntimeEnvVars verifies that
// CreateDefaultConfig returns a viper instance in which the
// defaults.runtimes.apptainer.* keys follow the
// MOAT_DEFAULTS_RUNTIMES_APPTAINER_* environment variables, even without
// going through InitConfig.
func TestCreateDefaultConfigApptainerRuntimeEnvVars(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_IMAGEURL", "ghcr.io/aaltorse/moat:env")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_CACHEDIR", "/tmp/moat_tests/env-cache")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_PASSENV", "false")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_MOUNTCWD", "true")

	cfg := CreateDefaultConfig()

	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	assert.Equal(t, "ghcr.io/aaltorse/moat:env", cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "/tmp/moat_tests/env-cache", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.False(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.mountcwd"))
}

// TestInitConfigApptainerRuntimeEnvVars verifies that the
// defaults.runtimes.apptainer.* keys follow the
// MOAT_DEFAULTS_RUNTIMES_APPTAINER_* environment variables (MOAT prefix,
// dots replaced by underscores).
func TestInitConfigApptainerRuntimeEnvVars(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	// Point the user home directory and working directory at empty
	// temporary directories so that the search paths cannot contain a
	// config file.
	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_IMAGEURL", "ghcr.io/aaltorse/moat:env")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_CACHEDIR", "/tmp/moat_tests/env-cache")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_PASSENV", "false")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_MOUNTCWD", "true")

	cfg, err := InitConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// The overridden keys hold the environment variable values.
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	assert.Equal(t, "ghcr.io/aaltorse/moat:env", cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "/tmp/moat_tests/env-cache", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.False(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.mountcwd"))
}

// TestInitConfigApptainerRuntimeTypeEnvVar verifies that
// MOAT_DEFAULTS_RUNTIMES_APPTAINER_TYPE overrides the default runtime type
// and that the environment-override configuration still passes validation.
func TestInitConfigApptainerRuntimeTypeEnvVar(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_TYPE", "custom")

	cfg, err := InitConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, "custom", cfg.GetString("defaults.runtimes.apptainer.type"))

	// The configuration with the environment override must still validate.
	require.NoError(t, ValidateConfig(cfg))
}

// TestInitConfigFileOverridesApptainerRuntimeEnvVars verifies that
// configuration file values take precedence over the
// MOAT_DEFAULTS_RUNTIMES_APPTAINER_* environment variable values: the
// environment variables override only the built-in defaults, not values
// given in the configuration file.
func TestInitConfigFileOverridesApptainerRuntimeEnvVars(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	// The environment variables give values that differ from the
	// configuration file and the built-in defaults.
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_IMAGEURL", "ghcr.io/aaltorse/moat:fromenv")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_CACHEDIR", "/tmp/moat_tests/fromenv")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_MOUNTCWD", "true")

	// The configuration file sets imageurl and cachedir, but not
	// mountcwd.
	cfgFile := filepath.Join(t.TempDir(), "moat-config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte(
		"defaults:\n"+
			"  runtime: apptainer\n"+
			"  runtimes:\n"+
			"    apptainer:\n"+
			"      type: apptainer\n"+
			"      imageurl: ghcr.io/aaltorse/moat:fromfile\n"+
			"      cachedir: /tmp/moat_tests/fromfile\n"+
			"      passenv: true\n"+
			"envs: {}\n"), 0o644))

	cfg, err := InitConfig(cfgFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// The configuration file values take precedence over the
	// environment variable values.
	assert.Equal(t, "apptainer", cfg.GetString("defaults.runtimes.apptainer.type"))
	assert.Equal(t, "ghcr.io/aaltorse/moat:fromfile", cfg.GetString("defaults.runtimes.apptainer.imageurl"))
	assert.Equal(t, "/tmp/moat_tests/fromfile", cfg.GetString("defaults.runtimes.apptainer.cachedir"))
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.passenv"))

	// A key the configuration file does not set still follows the
	// environment variable.
	assert.True(t, cfg.GetBool("defaults.runtimes.apptainer.mountcwd"))

	// The runtime spec lookup (used by moat run) also reflects the
	// configuration file values over the environment variable values.
	spec, err := GetRuntimeSpec(cfg, "apptainer")
	require.NoError(t, err)
	assert.Equal(t, "apptainer", spec.Type)
	assert.Equal(t, "ghcr.io/aaltorse/moat:fromfile", spec.ImageUrl)
	assert.Equal(t, "/tmp/moat_tests/fromfile", spec.CacheDir)
}

// TestGetRuntimeSpecApptainerEnvVars verifies that GetRuntimeSpec returns
// the apptainer runtime spec with the
// MOAT_DEFAULTS_RUNTIMES_APPTAINER_* environment variable overrides
// applied.
func TestGetRuntimeSpecApptainerEnvVars(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	t.Setenv("HOME", t.TempDir())
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { _ = os.Chdir(wd) })

	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_IMAGEURL", "ghcr.io/aaltorse/moat:env")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_CACHEDIR", "/tmp/moat_tests/env-cache")
	t.Setenv("MOAT_DEFAULTS_RUNTIMES_APPTAINER_MOUNTCWD", "true")

	cfg, err := InitConfig("")
	require.NoError(t, err)

	spec, err := GetRuntimeSpec(cfg, "apptainer")
	require.NoError(t, err)

	assert.Equal(t, "apptainer", spec.Type)
	assert.Equal(t, "ghcr.io/aaltorse/moat:env", spec.ImageUrl)
	assert.Equal(t, "/tmp/moat_tests/env-cache", spec.CacheDir)
	assert.True(t, spec.PassEnv)
	assert.True(t, spec.MountCWD)
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
