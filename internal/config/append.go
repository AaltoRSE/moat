package config

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// AppendConfig appends a value to a list config key and validates the
// result. It does not write the configuration to disk; callers that need
// to persist the change must call WriteConfig separately.
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

	return nil
}
