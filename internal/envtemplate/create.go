package envtemplate

import (
	"os"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

// CreateEnvTemplate creates a new environment template in the viper config.
func CreateEnvTemplate(name string, tmpl types.EnvTemplate) error {

	var err error

	templates := viper.GetStringMap("envtemplates")
	if templates[name] != nil {
		log.Error().Str("name", name).Msg("Environment template already exists.")
		return nil
	}

	if !utils.CheckEnvironmentName(name) {
		log.Error().Str("name", name).Msg("Invalid template name. Name must be non-empty and contain only alphanumeric characters and underscores.")
		return nil
	}

	// Sanitize and validate the home base directory
	tmpl.HomeBase, err = utils.SanitizeFolderPath(tmpl.HomeBase)
	if err != nil {
		log.Error().Err(err).Msg("Error getting absolute path for home base directory")
		return err
	}
	if !utils.CheckFolderExists(tmpl.HomeBase) {
		log.Info().Str("homeBase", tmpl.HomeBase).Msg("Home base directory does not exist")
		confirm := confirmation.New("Do you want to create the home base directory?", confirmation.Undecided)
		ready, err := confirm.RunPrompt()
		if err != nil {
			log.Error().Err(err).Msg("Error during confirmation prompt")
			return err
		}
		if ready {
			if err := os.MkdirAll(tmpl.HomeBase, 0755); err != nil {
				log.Error().Err(err).Msg("Failed to create home base directory")
				return err
			}
			log.Info().Str("homeBase", tmpl.HomeBase).Msg("Home base directory created")
		} else {
			log.Info().Msg("Aborting template creation due to missing home base directory.")
			return nil
		}
	}

	// Sanitize and validate the mounts

	tmpl.Mounts, err = utils.SanitizeMountsPaths(tmpl.Mounts)
	if err != nil {
		log.Error().Err(err).Msg("Error sanitizing mount paths")
		return err
	}

	if !utils.CheckMounts(tmpl.Mounts) {
		log.Error().Msg("One or more mount paths are invalid or do not exist.")
		return nil
	}

	log.Info().Str("name", name).Msg("Creating environment template")
	viper.Set("envtemplates."+name, tmpl)

	if err := config.WriteConfig(); err != nil {
		return err
	}

	log.Info().Str("name", name).Msg("Environment template created successfully.")
	return nil
}
