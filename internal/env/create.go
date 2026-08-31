package env

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/spf13/viper"
)

// CreateEnvironment creates a new environment configuration in the viper config.
func CreateEnvironment(name string, env types.MoatEnv) error {

	var envs = viper.GetStringMap("envs")

	if envs[name] != nil {
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
		// Ask to create the fake home directory if it doesn't exist
		log.Info().Str("absFakeHome", absFakeHome).Msg("Fake home directory does not exist")
		confirm := confirmation.New("Do you want to create the fake home directory?", confirmation.Undecided)
		ready, err := confirm.RunPrompt()
		if err != nil {
			log.Error().Err(err).Msg("Error during confirmation prompt")
			return err
		}
		if ready {
			err := os.MkdirAll(absFakeHome, 0755)
			if err != nil {
				log.Error().Err(err).Msg("Failed to create fake home directory")
				return err
			}
			log.Info().Str("absFakeHome", absFakeHome).Msg("Fake home directory created")
		} else {
			log.Info().Msg("Aborting environment creation due to missing fake home directory.")
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
		log.Error().Msg("One or more mount paths are invalid or do not exist.")
		return nil
	}

	// Create environment configuration
	fmt.Printf("Creating environment '%s'\n", name)
	viper.Set("envs."+name, env)

	// Write the updated configuration back to the config file
	if err := config.WriteConfig(); err != nil {
		return err
	}

	fmt.Println("Environment created successfully.")
	return nil
}
