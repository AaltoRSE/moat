package config

import (
	"os"
	"testing"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfigOverlayDefaults verifies that the default configuration
// registers the overlay settings for the apptainer runtime and passes
// validation.
func TestDefaultConfigOverlayDefaults(t *testing.T) {
	clearMoatDefaultEnvVars(t)

	cfg := CreateDefaultConfig()
	assert.Equal(t, "internal", cfg.GetString("defaults.runtimes.apptainer.overlayowner"))
	assert.Equal(t, "fuse-overlayfs", cfg.GetString("defaults.runtimes.apptainer.overlaydriver"))
	assert.Equal(t, "auto", cfg.GetString("defaults.runtimes.apptainer.overlayfssnapshot"))
	assert.Equal(t, "$HOME/.cache/moat/overlays", cfg.GetString("defaults.runtimes.apptainer.overlaydir"))
	// The docker default runtime has no overlay settings
	assert.Empty(t, cfg.GetString("defaults.runtimes.docker.overlayowner"))

	require.NoError(t, ValidateConfig(cfg))
}

// TestValidateRuntimeSpecOverlayRules verifies the cross-field validation
// rules of a runtime spec: overlay settings require the apptainer type on
// user runtimes, and the internal owner only supports fuse-overlayfs.
func TestValidateRuntimeSpecOverlayRules(t *testing.T) {
	base := types.RuntimeSpec{
		Type:     "apptainer",
		ImageUrl: "ghcr.io/aaltorse/moat:latest",
		CacheDir: "/tmp/moat_tests/cache",
		PassEnv:  true,
	}

	valid := []types.RuntimeSpec{
		base,
		func() types.RuntimeSpec { s := base; s.OverlayOwner = "internal"; return s }(),
		func() types.RuntimeSpec {
			s := base
			s.OverlayOwner = "internal"
			s.OverlayDriver = "fuse-overlayfs"
			return s
		}(),
		func() types.RuntimeSpec {
			s := base
			s.OverlayOwner = "external"
			s.OverlayDriver = "kernel-overlayfs"
			return s
		}(),
		func() types.RuntimeSpec {
			s := base
			s.OverlayOwner = "external"
			s.OverlayDriver = "btrfs"
			s.OverlayDir = "/tmp/moat_tests/overlays"
			return s
		}(),
	}
	for _, spec := range valid {
		require.NoError(t, ValidateRuntimeSpec(spec), "spec %+v should be valid", spec)
	}

	invalid := []types.RuntimeSpec{
		// internal owner with a non-fuse driver
		func() types.RuntimeSpec {
			s := base
			s.OverlayOwner = "internal"
			s.OverlayDriver = "kernel-overlayfs"
			return s
		}(),
		// driver set without owner is treated as internal
		func() types.RuntimeSpec { s := base; s.OverlayDriver = "btrfs"; return s }(),
		// overlay settings on a docker runtime
		func() types.RuntimeSpec {
			s := base
			s.Type = "docker"
			s.ImageUrl = ""
			s.CacheDir = ""
			s.OverlayOwner = "external"
			return s
		}(),
		func() types.RuntimeSpec {
			s := base
			s.Type = "docker"
			s.ImageUrl = ""
			s.CacheDir = ""
			s.OverlayDir = "/tmp/moat_tests/overlays"
			return s
		}(),
		// invalid enum values are rejected by the validator tags
		func() types.RuntimeSpec { s := base; s.OverlayOwner = "bogus"; return s }(),
		func() types.RuntimeSpec { s := base; s.OverlayDriver = "ext4"; return s }(),
	}
	for _, spec := range invalid {
		require.Error(t, ValidateRuntimeSpec(spec), "spec %+v should be invalid", spec)
	}
}

// TestValidateOverlayMounts verifies the format validation of overlay
// mount entries.
func TestValidateOverlayMounts(t *testing.T) {
	require.NoError(t, ValidateOverlayMounts(nil))
	require.NoError(t, ValidateOverlayMounts([]string{"/opt/conda"}))
	require.NoError(t, ValidateOverlayMounts([]string{"/opt/conda", "/usr/local", "/data:/host/data"}))
	t.Setenv("OVERLAY_PATH", "/expanded:/bad")
	require.Error(t, ValidateOverlayMounts([]string{"$OVERLAY_PATH"}))
	mounts, err := NormalizeOverlayMounts([]string{"~/overlays:/opt/${UID}"})
	require.NoError(t, err)
	assert.NotContains(t, mounts[0], "$")

	invalid := [][]string{
		{"opt/conda"}, // relative destination
		{"/opt", "/opt/conda"},
		{"/space in dest"},
		{"/comma,dest"},                     // mount option separator
		{"/"},                               // root destination
		{"/opt/../etc"},                     // .. element
		{"/opt/conda/"},                     // uncleaned destination
		{"/opt/conda", "/opt/conda"},        // duplicate destination
		{"/host/src:/data", "/other:/data"}, // duplicate destination with sources
		{"relsrc:/data"},                    // relative source
		{":/data"},                          // empty source
		{"/a:/b:/c"},                        // more than one colon
		{""},                                // empty entry
	}
	for _, mounts := range invalid {
		require.Error(t, ValidateOverlayMounts(mounts), "mounts %v should be invalid", mounts)
	}
}

// TestValidateConfigEnvOverlayRuntime verifies that an environment with
// overlay mounts fails validation when its resolved runtime is not of the
// apptainer type, and passes with the default apptainer runtime.
func TestValidateConfigEnvOverlayRuntime(t *testing.T) {
	home := t.TempDir()

	envWithOverlays := map[string]any{"home": home, "overlaymounts": []string{"/opt/conda"}}

	// The default runtime is apptainer: valid
	cfg := CreateDefaultConfig()
	cfg.Set("envs", map[string]any{"test": envWithOverlays})
	require.NoError(t, ValidateConfig(cfg))

	// Point the default runtime at docker: the environment with overlay
	// mounts must now fail validation
	cfg = CreateDefaultConfig()
	cfg.Set("envs", map[string]any{"test": envWithOverlays})
	cfg.Set("defaults.runtime", "docker")
	err := ValidateConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supported for the apptainer type")

	// An environment without overlay mounts is fine with the docker runtime
	cfg = CreateDefaultConfig()
	cfg.Set("envs", map[string]any{"test": map[string]any{"home": home}})
	cfg.Set("defaults.runtime", "docker")
	require.NoError(t, ValidateConfig(cfg))

	// An environment that explicitly selects the docker default runtime
	// with overlay mounts fails validation
	cfg = CreateDefaultConfig()
	cfg.Set("envs", map[string]any{"test": map[string]any{
		"home": home, "overlaymounts": []string{"/opt/conda"}, "runtime": "docker",
	}})
	err = ValidateConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "only supported for the apptainer type")

	// Invalid overlay mount entries fail validation
	cfg = CreateDefaultConfig()
	cfg.Set("envs", map[string]any{"test": map[string]any{
		"home": home, "overlaymounts": []string{"opt/conda"},
	}})
	err = ValidateConfig(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "overlay mounts")
}

// TestGetEnvSetsNameAndSanitizesOverlayMounts verifies that GetEnv sets the
// environment name and sanitizes the overlay mount paths when sanitized is
// true.
func TestGetEnvSetsNameAndSanitizesOverlayMounts(t *testing.T) {
	t.Setenv("MOAT_TEST_OVERLAY_SRC", t.TempDir())
	home := t.TempDir()

	cfg := CreateDefaultConfig()
	cfg.Set("envs", map[string]any{
		"myenv": map[string]any{
			"home":          home,
			"overlaymounts": []string{"/opt/conda", "$MOAT_TEST_OVERLAY_SRC:/data"},
		},
	})
	require.NoError(t, ValidateConfig(cfg))

	env, err := GetEnv(cfg, "myenv", true)
	require.NoError(t, err)
	assert.Equal(t, "myenv", env.Name)
	require.Len(t, env.OverlayMounts, 2)
	assert.Equal(t, "/opt/conda", env.OverlayMounts[0])
	assert.Equal(t, os.Getenv("MOAT_TEST_OVERLAY_SRC")+":/data", env.OverlayMounts[1])

	// Without sanitization the name is still set but the paths stay raw
	env, err = GetEnv(cfg, "myenv", false)
	require.NoError(t, err)
	assert.Equal(t, "myenv", env.Name)
	assert.Equal(t, "$MOAT_TEST_OVERLAY_SRC:/data", env.OverlayMounts[1])

	// GetEnvs sets the name of every environment
	envs := GetEnvs(cfg)
	assert.Equal(t, "myenv", envs["myenv"].Name)
}
