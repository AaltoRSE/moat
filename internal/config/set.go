package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// SetConfig sets a config key to value, validates, and persists to disk.
func SetConfig(cfg *viper.Viper, key string, value any) error {

	var (
		C   types.Config
		err error
	)

	v := viper.New()

	v.Set(key, value)

	err = cfg.MergeConfigMap(v.AllSettings())
	if err != nil {
		return fmt.Errorf("error when adding %q to config: %v", key, err)
	}

	if err = cfg.Unmarshal(&C); err != nil {
		return fmt.Errorf("invalid configuration after setting %q: %v", key, err)
	}

	if err := validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after setting %q: %v", key, err)
	}

	return WriteConfig(cfg)
}
