package env

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/spf13/viper"
)

// RemoveEnvironment removes an environment configuration from the viper config.
func RemoveEnvironment(name string) error {
	// Check if the environment exists
	envs := viper.GetStringMapString("envs")
	if _, ok := envs[name]; !ok {
		log.Error().Msgf("Environment does not exist: %s", name)
		return nil
	}

	delete(envs, name)
	viper.Set("envs", envs)
	if err := config.WriteConfig(); err != nil {
		return err
	}
	log.Info().Msgf("Environment removed successfully: %s", name)
	return nil
}
