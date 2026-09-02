// Package config manages moat's configuration. It wraps the global viper
// instance, providing initialization, validation, persistence, and lookup
// of configuration values.
package config

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	yaml "go.yaml.in/yaml/v3"
)

// InitConfig initializes the global viper configuration. It registers
// default values, loads the config file (named moat-config.yaml) from the
// given path or from the default search locations, unmarshals it into a
// types.Config, and validates the result.
func InitConfig(cfgFile string) (cfg *viper.Viper, err error) {

	// Set viper configuration instance
	cfg = viper.New()

	cfg.SetConfigName("moat-config")
	cfg.SetConfigType("yaml")
	if cfgFile != "" {
		cfg.SetConfigFile(cfgFile)
	}

	// Set defaults if not set
	cfg.SetDefault("defaults.runtimes.apptainer.type", "apptainer")
	cfg.SetDefault("defaults.runtimes.apptainer.imageurl", "ghcr.io/aaltorse/vscode-apptainer:latest")
	cfg.SetDefault("defaults.runtimes.apptainer.cachedir", "$HOME/.cache/moat/images")
	cfg.SetDefault("defaults.runtimes.apptainer.passenv", true)
	cfg.SetDefault("defaults.runtime", "apptainer")
	cfg.SetDefault("envs", map[string]types.MoatEnv{})

	cfg.AddConfigPath("$HOME/.config/moat")
	cfg.AddConfigPath(".")
	log.Debug().Msg("Reading configuration from file")
	err = cfg.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %v", err)
		} else {
			return nil, fmt.Errorf("config file not found")
		}
	}
	log.Debug().Msgf("Configuration loaded from file: %s", cfg.ConfigFileUsed())

	var config types.Config
	err = cfg.Unmarshal(&config)
	if err != nil {
		fmt.Println("Unable to unmarshal config", err)
		// Print the configuration as a string for debugging
		configStr := GetConfigAsString(cfg)
		log.Error().Msgf("Current configuration:\n%s", configStr)
		return nil, fmt.Errorf("unable to unmarshal config: %v", err)
	}

	if err := validateConfig(&config); err != nil {
		configStr := GetConfigAsString(cfg)
		log.Error().Msgf("Current configuration:\n%s", configStr)
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return cfg, nil
}

// WriteConfig writes the current viper configuration to the config file.
func WriteConfig(cfg *viper.Viper) error {
	err := cfg.WriteConfig()
	if err != nil {
		log.Error().Msgf("Error writing config: %v", err)
		return err
	}
	return nil
}

// validateConfig validates the configuration struct using the
// go-playground/validator tags. On failure it logs each validation error
// in detail and returns the error.
func validateConfig(config *types.Config) error {
	// Validate that the configuration struct is properly filled out
	var validate = validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(config)
	if err != nil {

		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
				log.Error().Msgf("Namespace: %s", e.Namespace())
				log.Error().Msgf("Field: %s", e.Field())
				log.Error().Msgf("StructNamespace: %s", e.StructNamespace())
				log.Error().Msgf("StructField: %s", e.StructField())
				log.Error().Msgf("Tag: %s", e.Tag())
				log.Error().Msgf("ActualTag: %s", e.ActualTag())
				log.Error().Msgf("Kind: %v", e.Kind())
				log.Error().Msgf("Type: %v", e.Type())
				log.Error().Msgf("Value: %v", e.Value())
				log.Error().Msgf("Param: %s", e.Param())
				log.Error().Msg("")
			}
		}
		return err
	}
	return nil
}

// GetConfigAsString returns the current viper configuration as a YAML
// string. It is intended for debugging and error reporting.
func GetConfigAsString(cfg *viper.Viper) string {
	c := cfg.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Err(err).Msg("unable to marshal config to YAML")
	}
	return string(bs)
}

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
			log.Error().Err(err).Msg("Failed to sanitize ReadOnly	Mounts paths")
			return types.MoatEnv{}, err
		}

		log.Debug().Interface("env", env).Msg("Sanitized environment configuration")
	}

	return env, err
}

func GetRuntimeSpec(cfg *viper.Viper, name string) (types.RuntimeSpec, error) {

	var (
		runtimeSpecMap map[string]any
		subConfig      *viper.Viper
		runtimeSpec    types.RuntimeSpec
		err            error
	)

	runtimeSpecMap = cfg.GetStringMap("runtimes." + name)
	if len(runtimeSpecMap) == 0 {
		subConfig = cfg.Sub("defaults.runtimes." + name)
		err = subConfig.Unmarshal(&runtimeSpec)
		if err != nil {
			return types.RuntimeSpec{}, fmt.Errorf("runtime not found")
		}
		log.Debug().Interface("runtimeSpec", runtimeSpec).Msg("Loaded runtime spec")
	} else {
		return types.RuntimeSpec{}, fmt.Errorf("runtimes %q not found", name)
	}

	return runtimeSpec, err
}

// GetVariableType returns the reflect.Type of the value stored in the
// viper configuration under the given key. It returns an error if the key
// does not exist in the configuration.
func GetVariableType(cfg *viper.Viper, key string) (reflect.Type, error) {
	if !cfg.IsSet(key) {
		return nil, fmt.Errorf("key %q not found in configuration", key)
	}
	value := cfg.Get(key)
	if value == nil {
		return nil, fmt.Errorf("value for key %q is nil", key)
	}
	return reflect.TypeOf(value), nil
}
