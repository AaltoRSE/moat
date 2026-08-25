package config

import (
	"fmt"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/viper"
)

// PrependConfig prepends a value to a list config key, validates, and persists to disk.
func PrependConfig(key, value string) error {
	current := viper.GetStringSlice(key)
	viper.Set(key, append([]string{value}, current...))

	var C types.Config
	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("invalid configuration after prepending to %q: %v", key, err)
	}

	if err := validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after prepending to %q: %v", key, err)
	}

	return WriteConfig()
}
