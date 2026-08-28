package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// AppendConfig appends a value to a list config key, validates, and persists to disk.
func AppendConfig(key, value string) error {
	current := viper.GetStringSlice(key)
	viper.Set(key, append(current, value))

	var C types.Config
	if err := viper.Unmarshal(&C); err != nil {
		return fmt.Errorf("invalid configuration after appending to %q: %v", key, err)
	}

	if err := validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after appending to %q: %v", key, err)
	}

	return WriteConfig()
}
