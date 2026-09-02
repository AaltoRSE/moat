package envtemplate

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"

	"github.com/AaltoRSE/moat/internal/config"
)

// RemoveEnvTemplate removes an environment template from the viper config.
// It returns an error if the template cannot be removed or the config cannot
// be written. If the template does not exist, it logs an error and returns
// nil.
func RemoveEnvTemplate(cfg *viper.Viper, name string) error {
	templates := cfg.GetStringMap("envtemplates")
	if _, ok := templates[name]; !ok {
		log.Error().Msgf("Environment template does not exist: %s", name)
		return nil
	}

	delete(templates, name)
	cfg.Set("envtemplates", templates)
	if err := config.WriteConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Environment template '%s' removed successfully\n", name)
	return nil
}
