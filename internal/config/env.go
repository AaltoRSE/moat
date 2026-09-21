package config

import (
	"fmt"

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
	log.Debug().Interface("env", env).Msg("Environment configuration")
	if sanitized {
		env.Home, err = utils.SanitizeFolderPath(env.Home)
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

		log.Debug().Interface("env", env).Msg("Sanitized environment configuration")
	}

	return env, err
}

// GetEnvs returns all configured environments as a map from environment
// name to MoatEnv. It returns an empty map if no environments are
// configured.
func GetEnvs(cfg *viper.Viper) map[string]types.MoatEnv {

	var envs map[string]types.MoatEnv

	envsCfg := cfg.Sub("envs")

	err := envsCfg.Unmarshal(&envs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal environments")
	}
	return envs
}
