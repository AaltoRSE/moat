package env

import (
	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/spf13/viper"
)

// CopyEnvironment copies an existing environment to a new environment.
//
// The new environment, named newName, starts as a copy of the source
// environment named sourceName. If home is non-empty, it overrides the
// source's home directory. If mounts is non-empty, it overrides the
// source's mounts. If command is non-empty, it overrides the source's
// command. All other fields of the source environment are preserved.
// autoCreate has the same meaning as in [CreateEnvironment].
func CopyEnvironment(cfg *viper.Viper, sourceName, newName, home string, mounts []string, command *string, autoCreate bool) error {
	source, err := config.GetEnv(cfg, sourceName, false)
	if err != nil {
		log.Error().Err(err).Msgf("could not get source environment %q", sourceName)
		return err
	}

	if home != "" {
		source.Home = home
	}
	if len(mounts) > 0 {
		source.Mounts = mounts
	}
	if command != nil {
		source.Command = command
	}

	return CreateEnvironment(cfg, newName, source, autoCreate)
}
