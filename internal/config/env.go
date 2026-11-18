package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/spf13/viper"
)

// GetEnv returns the named environment from the configuration. It returns
// an error if the environment does not exist or its fields cannot be
// unmarshaled into types.MoatEnv.
func GetEnv(cfg *viper.Viper, name string, sanitized bool) (types.MoatEnv, error) {

	envViper := cfg.Sub("envs." + name)

	var env types.MoatEnv

	if envViper == nil {
		log.Debug().Msgf("Environment %q not found", name)
		err := fmt.Errorf("environment %q not found", name)
		return types.MoatEnv{}, err
	}
	err := envViper.UnmarshalExact(&env)

	if err != nil {
		log.Debug().Msgf("Failed to unmarshal environment %q: %v", name, err)
		return types.MoatEnv{}, err
	}
	// The environment name is not stored in the configuration; it is set
	// here so that runtime engines can key per-environment state by it.
	env.Name = name
	log.Debug().Interface("env", env).Msg("Environment configuration")
	if sanitized {
		env.Home, err = utils.SanitizePath(env.Home)
		if err != nil {
			log.Error().Err(err).Msg("Failed to sanitize Home path")
			return types.MoatEnv{}, err
		}
		env.Mounts, err = utils.SanitizeMountsPaths(env.Mounts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to sanitize Mounts paths")
			return types.MoatEnv{}, err
		}
		env.ReadOnlyMounts, err = utils.SanitizeMountsPaths(env.ReadOnlyMounts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to sanitize ReadOnly Mounts paths")
			return types.MoatEnv{}, err
		}
		env.OverlayMounts, err = NormalizeOverlayMounts(env.OverlayMounts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to sanitize Overlay Mounts paths")
			return types.MoatEnv{}, err
		}

		log.Debug().Interface("env", env).Msg("Sanitized environment configuration")
	}

	return env, err
}

// GetEnvs returns all configured environments as a map from environment
// name to MoatEnv. The Name field of each environment is set to its map
// key. It returns an empty map if no environments are configured.
func GetEnvs(cfg *viper.Viper) map[string]types.MoatEnv {

	var envs map[string]types.MoatEnv

	envsCfg := cfg.Sub("envs")

	err := envsCfg.Unmarshal(&envs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal environments")
	}
	for name, env := range envs {
		env.Name = name
		envs[name] = env
	}
	return envs
}

// ValidateOverlayMounts checks the format and destinations of mounts.
// See [NormalizeOverlayMounts] for the path rules.
func ValidateOverlayMounts(mounts []string) error {
	_, err := NormalizeOverlayMounts(mounts)
	return err
}

// NormalizeOverlayMounts expands paths in [source:]destination entries.
// The source must be absolute after expansion. The destination must be a
// clean, absolute container path other than the root. A literal @cwd source
// is preserved for run time and may omit the destination. Destinations must
// be distinct and cannot be nested or contain mount option delimiters.
func NormalizeOverlayMounts(mounts []string) ([]string, error) {
	normalized := make([]string, len(mounts))
	destinations := make(map[string]bool, len(mounts))
	for i, mount := range mounts {
		source, dest, explicit := strings.Cut(mount, ":")
		if !explicit {
			if source == utils.CwdMountSource {
				if destinations[utils.CwdMountSource] {
					return nil, fmt.Errorf("duplicate overlay mount destination @cwd")
				}
				destinations[utils.CwdMountSource] = true
				normalized[i] = source
				continue
			}
			dest = source
		}
		dest, err := utils.ExpandPath(dest)
		if err != nil {
			return nil, fmt.Errorf("invalid overlay mount %q: %w", mount, err)
		}
		if !filepath.IsAbs(dest) || filepath.Clean(dest) != dest || dest == "/" {
			return nil, fmt.Errorf("invalid overlay mount %q: destination must be a clean absolute path other than /", mount)
		}
		if strings.ContainsAny(dest, ":,\\\x00") || strings.ContainsFunc(dest, unicode.IsSpace) {
			return nil, fmt.Errorf("invalid overlay mount %q: destination contains a mount delimiter or whitespace", mount)
		}
		if destinations[dest] {
			return nil, fmt.Errorf("duplicate overlay mount destination %q", dest)
		}
		destinations[dest] = true
		normalized[i] = dest
		if explicit {
			if source == utils.CwdMountSource {
				normalized[i] = source + ":" + dest
				continue
			}
			source, err = utils.ExpandPath(source)
			if err != nil {
				return nil, fmt.Errorf("invalid overlay mount %q: %w", mount, err)
			}
			if !filepath.IsAbs(source) || strings.ContainsAny(source, ":\r\n\x00") {
				return nil, fmt.Errorf("invalid overlay mount %q: source must be an absolute path without mount delimiters", mount)
			}
			normalized[i] = filepath.Clean(source) + ":" + dest
		}
	}
	for dest := range destinations {
		if dest == utils.CwdMountSource {
			continue
		}
		for parent := filepath.Dir(dest); parent != "/"; parent = filepath.Dir(parent) {
			if destinations[parent] {
				return nil, fmt.Errorf("overlay mount destination %q is nested under %q", dest, parent)
			}
		}
	}
	return normalized, nil
}

// ResolveEnvOverlayRuntime returns the Apptainer runtime selected by env.
// If env has no overlays, it returns a zero spec. It returns an error if
// overlays are configured and the selected runtime cannot provide them.
func ResolveEnvOverlayRuntime(cfg *viper.Viper, env types.MoatEnv) (types.RuntimeSpec, error) {
	if len(env.OverlayMounts) == 0 {
		return types.RuntimeSpec{}, nil
	}
	runtimeName := cfg.GetString("defaults.runtime")
	if env.Runtime != nil {
		runtimeName = *env.Runtime
	}
	spec, err := GetRuntimeSpec(cfg, runtimeName)
	if err != nil {
		return types.RuntimeSpec{}, fmt.Errorf("environment has overlay mounts but its runtime %q could not be resolved: %w", runtimeName, err)
	}
	if spec.Type != "apptainer" {
		return types.RuntimeSpec{}, fmt.Errorf("environment has overlay mounts but runtime %q is of type %q: overlay mounts are only supported for the apptainer type", runtimeName, spec.Type)
	}
	return spec, nil
}

// validateEnvOverlays validates the overlay mounts of a single environment:
// their entry format (ValidateOverlayMounts) and the type of the runtime
// the environment resolves to (ResolveEnvOverlayRuntime). It returns an error
// if either check fails.
func validateEnvOverlays(cfg *viper.Viper, name string, env types.MoatEnv) error {
	if err := ValidateOverlayMounts(env.OverlayMounts); err != nil {
		return fmt.Errorf("invalid overlay mounts of environment %q: %v", name, err)
	}
	if _, err := ResolveEnvOverlayRuntime(cfg, env); err != nil {
		return fmt.Errorf("environment %q: %v", name, err)
	}
	return nil
}
