// Package config manages moat's configuration. It wraps the global viper
// instance, providing initialization, validation, persistence, and lookup
// of configuration values.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/AaltoRSE/moat/internal/version"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	yaml "go.yaml.in/yaml/v3"
)

var CmdConfig *viper.Viper

// InitConfig initializes a viper configuration. It registers default
// values, loads the config file (named moat-config.yaml) from the given
// path or from the default search locations, unmarshals it into a
// types.Config, and validates the result. If no config file is found, no
// error is returned and the returned configuration holds the default
// configuration contents.
func InitConfig(cfgFile string) (cfg *viper.Viper, err error) {

	// Set viper configuration instance
	cfg = viper.New()

	cfg.SetConfigName("moat-config")
	cfg.SetConfigType("yaml")
	if cfgFile != "" {
		cfg.SetConfigFile(cfgFile)
	}

	// Set defaults if not set
	registerDefaults(cfg)

	configPath, err := defaultConfigPath()
	if err != nil {
		log.Error().Err(err).Msg("Failed to determine user's home directory")
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current working directory")
		return nil, err
	}
	cfg.AddConfigPath(cwd)
	cfg.AddConfigPath(filepath.Dir(configPath))
	log.Debug().Msg("Reading configuration from file")
	err = cfg.ReadInConfig()
	if err != nil {
		var configFileNotFoundErr viper.ConfigFileNotFoundError
		// An explicitly set config file that does not exist yields a plain
		// not-exist error, while a search-path miss yields a
		// ConfigFileNotFoundError. Both mean no configuration file is present.
		if errors.As(err, &configFileNotFoundErr) || os.IsNotExist(err) {
			log.Debug().Msg("Configuration file not found, using default configuration")
		} else {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
	} else {
		log.Debug().Msgf("Configuration loaded from file: %s", cfg.ConfigFileUsed())
	}

	if err := validateConfig(cfg); err != nil {
		// Print the configuration as a string for debugging
		configStr := GetConfigAsString(cfg)
		log.Error().Msgf("Current configuration:\n%s", configStr)
		return nil, fmt.Errorf("config validation failed: %v", err)
	}

	return cfg, nil
}

// registerDefaults sets the default moat configuration values on cfg.
func registerDefaults(cfg *viper.Viper) {

	var imageTag string

	if version.MoatVersion != "" {
		imageTag = strings.Split(version.MoatVersion, "-")[0]
	} else {
		imageTag = "latest"
	}

	cfg.SetDefault("defaults.runtimes.apptainer.type", "apptainer")
	cfg.SetDefault("defaults.runtimes.apptainer.imageurl", fmt.Sprintf("ghcr.io/aaltorse/moat:%s", imageTag))
	cfg.SetDefault("defaults.runtimes.apptainer.cachedir", "$HOME/.cache/moat/images")
	cfg.SetDefault("defaults.runtimes.apptainer.passenv", true)
	cfg.SetDefault("defaults.runtimes.apptainer.mountcwd", false)
	cfg.SetDefault("defaults.runtime", "apptainer")
	cfg.SetDefault("envs", map[string]types.MoatEnv{})
}

// defaultConfigPath returns the path of the global moat configuration
// file ($HOME/.config/moat/moat-config.yaml). It returns an error if the
// user's home directory cannot be determined.
func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user's home directory: %v", err)
	}
	return filepath.Join(home, ".config", "moat", "moat-config.yaml"), nil
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

// SetConfigPath switches the configuration file path of cfg to
// outputPath so that a subsequent [WriteConfig] call writes the
// configuration to outputPath, creating the parent directory when
// missing. When outputPath is empty, the global configuration path
// ($HOME/.config/moat/moat-config.yaml) is used. If a configuration file
// has already been found and no output path is given, the configuration
// is left as it is, an informational message is printed, and
// changed=false is reported. If the path of the found configuration is
// the same as outputPath, an error is returned. Otherwise the
// configuration file path is set via viper.SetConfigFile, the
// configuration is validated, an informational message is printed, and
// changed=true is reported. It returns an error if the configuration is
// invalid.
func SetConfigPath(cfg *viper.Viper, outputPath string) (changed bool, err error) {
	source := cfg.ConfigFileUsed()

	// A configuration file counts as found only when the reported path
	// actually exists; an explicitly given but missing file (e.g. via
	// the --config flag) has not been loaded.
	found := false
	if source != "" {
		if _, err := os.Stat(source); err == nil {
			found = true
		} else if !os.IsNotExist(err) {
			return false, fmt.Errorf("failed to access configuration file %s: %v", source, err)
		}
	}

	if outputPath == "" {
		if found {
			fmt.Printf("Configuration already exists: %s\n", source)
			return false, nil
		}
		outputPath, err = defaultConfigPath()
		if err != nil {
			return false, err
		}
	} else if found && outputPath == source {
		return false, fmt.Errorf("configuration source and output path are the same: %s", outputPath)
	}

	cfg.SetConfigFile(outputPath)

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return false, fmt.Errorf("failed to create configuration directory %s: %v", outputDir, err)
	}

	// Validate the configuration, similar to SetConfig.
	if err := validateConfig(cfg); err != nil {
		return false, fmt.Errorf("config validation failed: %v", err)
	}

	log.Debug().Str("path", outputPath).Msg("Set configuration file path")
	fmt.Printf("Initialized configuration: %s\n", outputPath)

	return true, nil
}

// validateConfig validates the configuration held in cfg. It unmarshals
// the viper instance into a types.Config and validates the result using
// the go-playground/validator tags. On validation failure it logs each
// validation error in detail. It returns an error if the configuration
// cannot be unmarshaled or fails validation.
func validateConfig(cfg *viper.Viper) error {
	var config types.Config
	if err := cfg.Unmarshal(&config); err != nil {
		return fmt.Errorf("unable to unmarshal config: %v", err)
	}

	// Validate that the configuration struct is properly filled out
	var validate = validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&config)
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
			log.Error().Err(err).Msg("Failed to sanitize ReadOnly Mounts paths")
			return types.MoatEnv{}, err
		}

		log.Debug().Interface("env", env).Msg("Sanitized environment configuration")
	}

	return env, err
}

func GetEnvs(cfg *viper.Viper) map[string]types.MoatEnv {

	var envs map[string]types.MoatEnv

	envsCfg := cfg.Sub("envs")

	err := envsCfg.Unmarshal(&envs)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal environments")
	}
	return envs
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

// GetConfigFile returns the path of the currently active configuration
// file. If cfg is nil or no configuration file has been loaded yet, it
// returns the path of the global moat-config.yaml
// ($HOME/.config/moat/moat-config.yaml). It returns an empty string if the
// global path cannot be determined.
func GetConfigFile(cfg *viper.Viper) string {
	if cfg != nil {
		if path := cfg.ConfigFileUsed(); path != "" {
			return path
		}
	}
	configPath, err := defaultConfigPath()
	if err != nil {
		log.Error().Err(err).Msg("Failed to determine user's home directory")
		return ""
	}
	return configPath
}

// GetVariableType returns the reflect.Type of the value stored under the
// given dot-separated key in the types.Config structure. The type is
// derived from the field definitions of types.Config, so it is also
// available for optional fields that have not been set in the current
// configuration.
//
// A segment that addresses a map field (envs, runtimes, defaults.runtimes)
// is treated as the map key and may have any value; the following segments
// are resolved against the map's element type. Pointer fields are reported
// as their element type.
//
// It returns an error if the key does not address a field of types.Config.
func GetVariableType(key string) (reflect.Type, error) {
	segments := strings.Split(key, ".")
	current := reflect.TypeOf(types.Config{})

	for i, segment := range segments {
		switch current.Kind() {
		case reflect.Struct:
			field, ok := findFieldByYAMLName(current, segment)
			if !ok {
				return nil, fmt.Errorf("key %q not found in configuration", strings.Join(segments[:i+1], "."))
			}
			current = field
		case reflect.Map:
			// A map key may have any value, so the segment is consumed
			// without a lookup and the element type becomes current.
			current = current.Elem()
		default:
			return nil, fmt.Errorf("key %q not found in configuration", strings.Join(segments[:i+1], "."))
		}
		current = stripPointers(current)
	}

	return current, nil
}

// findFieldByYAMLName returns the type of the field of the struct type t
// whose yaml tag name (or field name, if the field has no yaml tag) matches
// name case-insensitively. The boolean result reports whether such a field
// exists.
func findFieldByYAMLName(t reflect.Type, name string) (reflect.Type, bool) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if strings.EqualFold(tag, name) || strings.EqualFold(field.Name, name) {
			return field.Type, true
		}
	}
	return nil, false
}

// stripPointers returns t with all levels of pointer indirection removed.
func stripPointers(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
