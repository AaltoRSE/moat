package config

import (
	"fmt"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/viper"
)

// SetConfig sets a config key to value, validates, and persists to disk.
func SetConfig(key, value string) error {
	viper.Set(key, value)

	var C types.Config
	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("invalid configuration after setting %q: %v", key, err)
	}

	if err := validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after setting %q: %v", key, err)
	}

	return WriteConfig()
}
