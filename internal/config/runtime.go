package config

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// ValidateRuntimeSpec validates a RuntimeSpec using the
// go-playground/validator tags of types.RuntimeSpec. On validation failure
// it logs each validation error in detail. It returns an error if the spec
// fails validation.
func ValidateRuntimeSpec(spec types.RuntimeSpec) error {
	var validate = validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(&spec)
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

// GetRuntimes returns all available runtimes as a map from runtime name to
// RuntimeSpec. It merges the default runtimes from defaults.runtimes with
// the user-specified runtimes from the top-level runtimes key; a
// user-specified runtime with the same name as a default runtime overwrites
// the default runtime. It returns an empty map if no runtimes are
// configured.
func GetRuntimes(cfg *viper.Viper) map[string]types.RuntimeSpec {

	var defaultRuntimes map[string]types.RuntimeSpec

	defaultRuntimesCfg := cfg.Sub("defaults.runtimes")
	if defaultRuntimesCfg != nil {
		if err := defaultRuntimesCfg.Unmarshal(&defaultRuntimes); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal default runtimes")
		}
	}

	userRuntimes := GetUserRuntimes(cfg)

	runtimes := make(map[string]types.RuntimeSpec, len(defaultRuntimes)+len(userRuntimes))
	for name, spec := range defaultRuntimes {
		runtimes[name] = spec
	}
	// User-specified runtimes overwrite default runtimes with the same name
	for name, spec := range userRuntimes {
		runtimes[name] = spec
	}
	return runtimes
}

// GetUserRuntimes returns the user-specified runtimes from the top-level
// runtimes key as a map from runtime name to RuntimeSpec. Unlike
// [GetRuntimes], it does not merge in the default runtimes from
// defaults.runtimes. It returns an empty map if no user-specified
// runtimes are configured.
func GetUserRuntimes(cfg *viper.Viper) map[string]types.RuntimeSpec {

	runtimes := make(map[string]types.RuntimeSpec)

	runtimesCfg := cfg.Sub("runtimes")
	if runtimesCfg == nil {
		return runtimes
	}

	if err := runtimesCfg.Unmarshal(&runtimes); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal user-specified runtimes")
		return make(map[string]types.RuntimeSpec)
	}

	return runtimes
}

// GetRuntimeSpec returns the runtime spec for the named runtime. A
// user-specified runtime from the top-level runtimes key takes precedence
// over a default runtime from defaults.runtimes with the same name. It
// returns an error if no runtime with the given name exists or its fields
// cannot be unmarshaled into types.RuntimeSpec.
func GetRuntimeSpec(cfg *viper.Viper, name string) (types.RuntimeSpec, error) {
	var runtimeSpec types.RuntimeSpec

	// A user-specified runtime takes precedence over the default runtimes
	runtimesCfg := cfg.Sub("runtimes." + name)
	if runtimesCfg != nil {
		if err := runtimesCfg.Unmarshal(&runtimeSpec); err != nil {
			return types.RuntimeSpec{}, fmt.Errorf("failed to unmarshal runtime %q: %v", name, err)
		}
		if runtimeSpec.Type != "" {
			log.Debug().Interface("runtimeSpec", runtimeSpec).Msg("Loaded runtime spec")
			return runtimeSpec, nil
		}
	}

	defaultsCfg := cfg.Sub("defaults.runtimes." + name)
	if defaultsCfg == nil {
		return types.RuntimeSpec{}, fmt.Errorf("runtime %q not found", name)
	}
	if err := defaultsCfg.Unmarshal(&runtimeSpec); err != nil {
		return types.RuntimeSpec{}, fmt.Errorf("failed to unmarshal runtime %q: %v", name, err)
	}
	if runtimeSpec.Type == "" {
		return types.RuntimeSpec{}, fmt.Errorf("runtime %q not found", name)
	}
	log.Debug().Interface("runtimeSpec", runtimeSpec).Msg("Loaded runtime spec")
	return runtimeSpec, nil
}
