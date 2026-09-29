package runtimes

import (
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// SetRuntime changes the fields of an existing runtime named name and
// writes the updated configuration to the config file.
//
// The runtime must already exist, either as a user-specified runtime in
// the top-level runtimes key or as a default runtime in
// defaults.runtimes. updates holds the fields to change, paired with
// their new values, in the order they should be applied; only the given
// fields are changed, and all other fields keep their current values.
// When the named runtime is a default runtime, a user-specified runtime
// with the same name is created from the default runtime's specification
// with the given changes applied, so that it overwrites the default
// runtime in runtime lookups; the default runtimes themselves are never
// modified. The cache directory is resolved to an absolute path before
// storing.
//
// It prints "Set {name}.{field} = {value}" for every applied field and a
// final success message. It returns an error if the runtime does not
// exist, an update names an unknown field, an update has the wrong number
// of values or an invalid value, the specification or the full
// configuration fail validation after the change, or the configuration
// cannot be written.
func SetRuntime(cfg *viper.Viper, name string, updates []types.RuntimeUpdate) error {

	if len(updates) == 0 {
		return fmt.Errorf("no fields to set for runtime %q", name)
	}

	// The runtime must already exist, either user-specified or default
	spec, err := config.GetRuntimeSpec(cfg, name)
	if err != nil {
		return err
	}

	userRuntimes := config.GetUserRuntimes(cfg)
	_, isUser := userRuntimes[name]

	// A default runtime is overridden by a user runtime seeded from the
	// default specification; resolve the seeded cache directory to an
	// absolute path, as CreateRuntime does for user runtimes
	if !isUser && spec.CacheDir != "" {
		absCacheDir, err := utils.SanitizeFolderPath(spec.CacheDir)
		if err != nil {
			log.Error().Err(err).Msg("Error getting absolute path for cache directory")
			return err
		}
		spec.CacheDir = absCacheDir
	}

	// Apply the updates to the current specification, in the given order
	applied := make([]string, 0, len(updates))
	for _, u := range updates {
		message, err := applyRuntimeUpdate(&spec, name, u)
		if err != nil {
			return err
		}
		applied = append(applied, message)
	}

	// Validate the updated specification before storing it
	if err := config.ValidateRuntimeSpec(spec); err != nil {
		log.Error().Err(err).Msg("Invalid runtime spec")
		return fmt.Errorf("invalid runtime spec: %v", err)
	}

	if !isUser {
		fmt.Printf("Overriding default runtime '%s' with a new user runtime\n", name)
	}
	for _, message := range applied {
		fmt.Println(message)
	}

	userRuntimes[name] = spec
	cfg.Set("runtimes", userRuntimes)

	// Validate the configuration before writing it to disk
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed after setting runtime %q: %v", name, err)
	}

	// Write the updated configuration back to the config file
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Runtime updated successfully.")
	return nil
}

// applyRuntimeUpdate applies a single RuntimeUpdate to the runtime spec
// and returns the "Set {name}.{field} = {value}" message for the applied
// change. The value printed for the cachedir field is the resolved
// absolute path. It returns an error if the update names an unknown
// field, has the wrong number of values, or holds an invalid value.
func applyRuntimeUpdate(spec *types.RuntimeSpec, name string, update types.RuntimeUpdate) (string, error) {

	var applied string

	switch update.Field {
	case "type":
		value, err := singleRuntimeUpdateValue(update)
		if err != nil {
			return "", err
		}
		spec.Type = value
		applied = value
	case "imageurl":
		value, err := singleRuntimeUpdateValue(update)
		if err != nil {
			return "", err
		}
		spec.ImageUrl = value
		applied = value
	case "cachedir":
		value, err := singleRuntimeUpdateValue(update)
		if err != nil {
			return "", err
		}
		absValue, err := utils.SanitizeFolderPath(value)
		if err != nil {
			log.Error().Err(err).Msg("Error getting absolute path for cache directory")
			return "", err
		}
		spec.CacheDir = absValue
		applied = absValue
	case "passenv":
		value, err := singleRuntimeUpdateValue(update)
		if err != nil {
			return "", err
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return "", fmt.Errorf("invalid boolean value %q for field %q: %v", value, update.Field, err)
		}
		spec.PassEnv = parsed
		applied = value
	case "mountcwd":
		value, err := singleRuntimeUpdateValue(update)
		if err != nil {
			return "", err
		}
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return "", fmt.Errorf("invalid boolean value %q for field %q: %v", value, update.Field, err)
		}
		spec.MountCWD = parsed
		applied = value
	default:
		return "", fmt.Errorf("unknown runtime field %q", update.Field)
	}

	return fmt.Sprintf("Set %s.%s = [%s]", name, update.Field, applied), nil
}

// singleRuntimeUpdateValue returns the single value of a runtime update,
// enforcing that the field receives exactly one value.
func singleRuntimeUpdateValue(update types.RuntimeUpdate) (string, error) {
	if len(update.Value) != 1 {
		return "", fmt.Errorf("field %q expects exactly 1 value, got %d", update.Field, len(update.Value))
	}
	return update.Value[0], nil
}
