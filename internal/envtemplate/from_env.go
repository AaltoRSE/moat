package envtemplate

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// CreateEnvTemplateFromEnv converts an existing environment into a new
// environment template and stores it in the configuration. It looks up the
// environment named envName, derives a template from its settings (using the
// environment's home directory as the template's home base), and persists the
// template under templateName. Any non-empty fields in overrides replace the
// corresponding values derived from the environment. It returns an error if the
// environment does not exist or the template cannot be created.
func CreateEnvTemplateFromEnv(cfg *viper.Viper, envName string, templateName string, overrides types.EnvTemplate) error {
	env, err := config.GetEnv(cfg, envName, false)
	if err != nil {
		log.Error().Err(err).Msg("Error retrieving environment")
		return err
	}

	tmpl := types.EnvTemplate{
		HomeBase:       env.Home,
		Mounts:         env.Mounts,
		ReadOnlyMounts: env.ReadOnlyMounts,
		Runtime:        env.Runtime,
		PassEnv:        env.PassEnv,
	}

	// Apply overrides where provided
	if overrides.HomeBase != "" {
		tmpl.HomeBase = overrides.HomeBase
	}
	if len(overrides.Mounts) > 0 {
		tmpl.Mounts = overrides.Mounts
	}

	return CreateEnvTemplate(cfg, templateName, tmpl)
}
