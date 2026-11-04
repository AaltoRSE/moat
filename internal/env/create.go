package env

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/prompts"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/spf13/viper"
)

// CreateEnvironment creates a new environment configuration in the viper config.
//
// If autoCreate is true, missing directories (the fake home directory and
// mount source paths) are created automatically. Otherwise, the user is
// prompted to create the missing directories; answering no, or an invalid
// mount path, aborts the creation.
func CreateEnvironment(cfg *viper.Viper, name string, env types.MoatEnv, autoCreate bool) error {

	// Get existing environments from the configuration
	var envs = config.GetEnvs(cfg)

	if _, exists := envs[name]; exists {
		fmt.Println("Environment already exists.")
		return nil
	}

	// Validate the environment name
	if !utils.CheckEnvironmentName(name) {
		log.Error().Str("name", name).Msg("Invalid environment name. Environment name must be non-empty and contain only alphanumeric characters, underscores, and hyphens.")
		return nil
	}

	// Sanitize the mount paths: expand a leading ~ and environment
	// variables, and resolve relative paths to absolute ones
	var err error
	env.Mounts, err = utils.SanitizeMountsPaths(env.Mounts)
	if err != nil {
		log.Error().Err(err).Msg("Error sanitizing mount paths")
		return err
	}

	// Sanitize the read-only mount paths the same way
	env.ReadOnlyMounts, err = utils.SanitizeMountsPaths(env.ReadOnlyMounts)
	if err != nil {
		log.Error().Err(err).Msg("Error sanitizing read-only mount paths")
		return err
	}

	// Ensure that the fake home directory and the mount source directories
	// exist, creating missing ones automatically or after asking the user
	if err := prompts.CheckAndCreateMounts(env, autoCreate); err != nil {
		log.Error().Err(err).Msg("Error ensuring environment directories exist")
		return err
	}

	// Create environment configuration
	fmt.Printf("Creating environment '%s'\n", name)

	envs[name] = env
	cfg.Set("envs", envs)

	// Write the updated configuration back to the config file
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}

	fmt.Println("Environment created successfully.")
	return nil
}
