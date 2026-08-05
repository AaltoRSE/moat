package env

import (
	"log"

	"github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/viper"
)

// RemoveEnvironment removes an environment configuration from the viper config.
func RemoveEnvironment(env types.SharkEnv) error {
	// Check if the environment exists
	envs := viper.GetStringMapString("envs")
	if _, ok := envs[env.Name]; !ok {
		log.Println("Environment does not exist:", env.Name)
		return nil
	}

	delete(envs, env.Name)
	viper.Set("envs", envs)
	config.WriteConfig()
	log.Println("Environment removed successfully:", env.Name)
	return nil
}
