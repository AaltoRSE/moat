// Package runtime provides functions for managing runtimes in the moat
// configuration.
//
// The functions in this package create and update user-specified runtimes
// and persist the changes to the config file; they are used by the moat
// runtime command group. Executing commands inside environments is
// handled by the runtime engines in the engines package.
package runtime

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// CreateRuntime creates a new user-specified runtime named name with the
// given spec in the top-level runtimes key of the configuration and writes
// the updated configuration to the config file.
//
// The name must be non-empty and contain only alphanumeric characters and
// underscores. If a user-specified runtime with the same name already
// exists, the configuration is left unchanged and no error is returned.
// Creating a runtime with the name of a default runtime is allowed; the
// new runtime then overwrites the default runtime in runtime lookups.
//
// It returns an error if the name is invalid, the spec fails validation,
// the cache directory cannot be resolved to an absolute path, the
// configuration fails validation after the change, or the configuration
// cannot be written.
func CreateRuntime(cfg *viper.Viper, name string, spec types.RuntimeSpec) error {

	// Validate the runtime name
	if !utils.CheckEnvironmentName(name) {
		log.Error().Str("name", name).Msg("Invalid runtime name. Runtime name must be non-empty and contain only alphanumeric characters and underscores.")
		return fmt.Errorf("invalid runtime name %q: name must be non-empty and contain only alphanumeric characters and underscores", name)
	}

	// Validate the runtime spec
	if err := config.ValidateRuntimeSpec(spec); err != nil {
		log.Error().Err(err).Msg("Invalid runtime spec")
		return fmt.Errorf("invalid runtime spec: %v", err)
	}

	runtimes := config.GetUserRuntimes(cfg)
	if _, exists := runtimes[name]; exists {
		fmt.Println("Runtime already exists.")
		return nil
	}

	// Resolve the cache directory to an absolute path before storing
	if spec.CacheDir != "" {
		absCacheDir, err := utils.SanitizeFolderPath(spec.CacheDir)
		if err != nil {
			log.Error().Err(err).Msg("Error getting absolute path for cache directory")
			return err
		}
		spec.CacheDir = absCacheDir
	}

	// Create the runtime configuration
	fmt.Printf("Creating runtime '%s'\n", name)

	runtimes[name] = spec
	cfg.Set("runtimes", runtimes)

	// Validate the configuration before writing it to disk
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed after creating runtime %q: %v", name, err)
	}

	// Write the updated configuration back to the config file
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Runtime created successfully.")
	return nil
}
