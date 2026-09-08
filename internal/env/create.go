package env

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/spf13/viper"

	"github.com/erikgeiser/promptkit/confirmation"
)

// CreateEnvironment creates a new environment configuration in the viper config.
//
// If autoCreate is true, missing directories (the fake home directory and
// mount source paths) are created automatically. Otherwise, the user is
// prompted to create a missing fake home directory, and missing or invalid
// mount paths abort the creation.
func CreateEnvironment(cfg *viper.Viper, name string, env types.MoatEnv, autoCreate bool) error {

	// Get existing environments from the configuration
	var envs = config.GetEnvs(cfg)

	if _, exists := envs[name]; exists {
		fmt.Println("Environment already exists.")
		return nil
	}

	// Validate the environment name
	if !utils.CheckEnvironmentName(name) {
		log.Error().Str("name", name).Msg("Invalid environment name. Environment name must be non-empty and contain only alphanumeric characters and underscores.")
		return nil
	}

	// Get the absolute path of the fake home directory
	absFakeHome, err := utils.SanitizeFolderPath(env.Home)
	if err != nil {
		log.Error().Err(err).Msg("Error getting absolute path for fake home directory")
		return err
	}

	// Validate that the fake home directory exists
	if !utils.CheckFolderExists(absFakeHome) {
		log.Info().Str("absFakeHome", absFakeHome).Msg("Fake home directory does not exist")
		var createHome bool
		if autoCreate {
			createHome = true
		} else {
			// Ask to create the fake home directory if it doesn't exist
			confirm := confirmation.New("Do you want to create the fake home directory?", confirmation.Undecided)
			ready, err := confirm.RunPrompt()
			if err != nil {
				log.Error().Err(err).Msg("Error during confirmation prompt")
				return err
			}
			createHome = ready
		}
		if createHome {
			if err := os.MkdirAll(absFakeHome, 0755); err != nil {
				log.Error().Err(err).Msg("Failed to create fake home directory")
				return err
			}
			fmt.Printf("Fake home directory created: %s\n", absFakeHome)
		} else {
			fmt.Println("Aborting environment creation due to missing fake home directory.")
			return nil
		}
	}

	// Validate that all project mounts exist
	env.Mounts, err = utils.SanitizeMountsPaths(env.Mounts)
	if err != nil {
		log.Error().Err(err).Msg("Error sanitizing mount paths")
		return err
	}
	if !utils.CheckMounts(env.Mounts) {
		if !autoCreate {
			log.Error().Msg("One or more mount paths are invalid or do not exist.")
			return nil
		}
		if err := utils.CreateMountDirs(env.Mounts); err != nil {
			return err
		}
		// Re-validate in case some mount entries are invalid rather than merely missing.
		if !utils.CheckMounts(env.Mounts) {
			log.Error().Msg("One or more mount paths are invalid.")
			return nil
		}
	}

	// Validate that all read-only project mounts exist
	env.ReadOnlyMounts, err = utils.SanitizeMountsPaths(env.ReadOnlyMounts)
	if err != nil {
		log.Error().Err(err).Msg("Error sanitizing read-only mount paths")
		return err
	}
	if !utils.CheckMounts(env.ReadOnlyMounts) {
		if !autoCreate {
			log.Error().Msg("One or more read-only mount paths are invalid or do not exist.")
			return nil
		}
		if err := utils.CreateMountDirs(env.ReadOnlyMounts); err != nil {
			return err
		}
		// Re-validate in case some mount entries are invalid rather than merely missing.
		if !utils.CheckMounts(env.ReadOnlyMounts) {
			log.Error().Msg("One or more read-only mount paths are invalid.")
			return nil
		}
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
