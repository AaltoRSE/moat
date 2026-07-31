package env

import (
	"log"

	config "github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/spf13/viper"
)

// RemoveEnvironment removes an environment configuration from the viper config.
func RemoveEnvironment(name string) error {
	// Check if the environment exists
	envs := viper.GetStringMapString("envs")
	if _, ok := envs[name]; !ok {
		log.Println("Environment does not exist:", name)
		return nil
	}

	delete(envs, name)
	viper.Set("envs", envs)
	config.WriteConfig()
	log.Println("Environment removed successfully:", name)
	return nil
}
