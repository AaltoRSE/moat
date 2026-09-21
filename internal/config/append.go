package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppendConfig appends a value to a list config key and validates the
// result. It does not write the configuration to disk; callers that need
// to persist the change must call WriteConfig separately.
func AppendConfig(cfg *viper.Viper, key, value string) error {
	current := cfg.GetStringSlice(key)
	cfg.Set(key, append(current, value))

	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed after appending to %q: %v", key, err)
	}

	return nil
}
