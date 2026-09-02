package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// AppendConfig appends a value to a list config key, validates, and persists to disk.
func AppendConfig(cfg *viper.Viper, key, value string) error {
	current := cfg.GetStringSlice(key)
	cfg.Set(key, append(current, value))

	var C types.Config
	if err := cfg.Unmarshal(&C); err != nil {
		return fmt.Errorf("invalid configuration after appending to %q: %v", key, err)
	}

	if err := validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after appending to %q: %v", key, err)
	}

	return WriteConfig(cfg)
}
