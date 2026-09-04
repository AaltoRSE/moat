package env

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/viper"
)

// RemoveEnvironment removes an environment configuration from the viper config.
func RemoveEnvironment(cfg *viper.Viper, name string) error {
	// Get existing environments from the configuration
	var envs = config.GetEnvs(cfg)

	// Check if the environment exists
	if _, ok := envs[name]; !ok {
		log.Error().Msgf("Environment does not exist: %s", name)
		return nil
	}

	delete(envs, name)
	cfg.Set("envs", envs)
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Environment removed successfully: %s\n", name)
	return nil
}
