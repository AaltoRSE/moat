package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// PrependConfig prepends a value to a list config key and validates the
// result. It does not write the configuration to disk; callers that need
// to persist the change must call WriteConfig separately.
func PrependConfig(cfg *viper.Viper, key, value string) error {
	current := cfg.GetStringSlice(key)
	cfg.Set(key, append([]string{value}, current...))

	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("config validation failed after prepending to %q: %v", key, err)
	}

	return nil
}
