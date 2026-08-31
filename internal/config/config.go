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
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	yaml "go.yaml.in/yaml/v3"
)

// InitConfig initializes the global viper configuration. It registers
// default values, loads the config file (named moat-config.yaml) from the
// given path or from the default search locations, unmarshals it into a
// types.Config, and validates the result.
func InitConfig(cfgFile string) error {
	viper.SetConfigName("moat-config")
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	// Set defaults if not set
	viper.SetDefault("defaults.runtime", "apptainer")
	viper.SetDefault("defaults.runtimes.apptainer.type", "apptainer")
	viper.SetDefault("defaults.runtimes.apptainer.imageurl", "ghcr.io/aaltorse/vscode-apptainer:latest")
	viper.SetDefault("defaults.runtimes.apptainer.cachedir", "$HOME/.cache/moat/images")
	viper.SetDefault("defaults.runtimes.apptainer.passenv", true)
	viper.SetDefault("envs", map[string]types.MoatEnv{})

	viper.AddConfigPath("$HOME/.config/moat")
	viper.AddConfigPath(".")
	log.Debug().Msg("Reading configuration from file")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %v", err)
		} else {
			return fmt.Errorf("config file not found")
		}
	}
	log.Debug().Msgf("Configuration loaded from file: %s", viper.ConfigFileUsed())

	var C types.Config
	err = viper.Unmarshal(&C)
	if err != nil {
		fmt.Println("Unable to unmarshal config", err)
		// Print the configuration as a string for debugging
		configStr := GetConfigAsString()
		log.Error().Msgf("Current configuration:\n%s", configStr)
		return fmt.Errorf("unable to unmarshal config: %v", err)
	}

	if err := validateConfig(&C); err != nil {
		configStr := GetConfigAsString()
		log.Error().Msgf("Current configuration:\n%s", configStr)
		return fmt.Errorf("config validation failed: %v", err)
	}

	return nil
}

// WriteConfig writes the current viper configuration to the config file.
func WriteConfig() error {
	err := viper.WriteConfig()
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
func GetConfigAsString() string {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Err(err).Msg("unable to marshal config to YAML")
	}
	return string(bs)
}

// GetEnv returns the named environment from the configuration. It returns
// an error if the environment does not exist or its fields cannot be
// unmarshaled into types.MoatEnv.
func GetEnv(name string) (types.MoatEnv, error) {

	envViper := viper.Sub("envs." + name)

	var env types.MoatEnv

	if envViper == nil {
		log.Debug().Msgf("Environment %q not found", name)
		err := fmt.Errorf("environment %q not found", name)
		return types.MoatEnv{}, err
	}
	err := envViper.UnmarshalExact(&env)

	if err != nil {
		log.Debug().Msgf("Failed to unmarshal environment %q: %v", name, err)
	} else {
		log.Debug().Interface("env", env).Msg("Environment configuration")
	}

	return env, err
}

// GetVariableType returns the reflect.Type of the value stored in the
// viper configuration under the given key. It returns an error if the key
// does not exist in the configuration.
func GetVariableType(key string) (reflect.Type, error) {
	if !viper.IsSet(key) {
		return nil, fmt.Errorf("key %q not found in configuration", key)
	}
	value := viper.Get(key)
	if value == nil {
		return nil, fmt.Errorf("value for key %q is nil", key)
	}
	return reflect.TypeOf(value), nil
}
