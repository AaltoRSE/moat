package config

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// SetConfig sets the config key to a value and validates the result. It
// does not write the configuration to disk; callers that need to persist
// the change must call WriteConfig separately. The value is interpreted
// according to the type of the key in types.Config: string and boolean
// keys expect exactly one element, while slice keys consume all elements.
// It returns the value that was set. It returns an error if the key is
// not a field of types.Config, the number of elements does not match the
// key type, an element is invalid, or the configuration is invalid after
// the change.
func SetConfig(cfg *viper.Viper, key string, value []string) error {

	keyType, err := GetVariableType(key)
	if err != nil {
		return err
	}

	var typedValue any

	switch keyType.Kind() {
	case reflect.String:
		if len(value) != 1 {
			return fmt.Errorf("key %q of type string expects exactly 1 value, got %d", key, len(value))
		}
		typedValue = value[0]
	case reflect.Bool:
		if len(value) != 1 {
			return fmt.Errorf("key %q of type bool expects exactly 1 value, got %d", key, len(value))
		}
		parsed, err := strconv.ParseBool(value[0])
		if err != nil {
			return fmt.Errorf("invalid boolean value %q for key %q: %v", value[0], key, err)
		}
		typedValue = parsed
	case reflect.Slice:
		typedValue = value
	default:
		return fmt.Errorf("unsupported key type for key %q: %v", key, keyType.Kind())
	}

	var (
		C types.Config
		v = viper.New()
	)

	v.Set(key, typedValue)

	err = cfg.MergeConfigMap(v.AllSettings())
	if err != nil {
		return fmt.Errorf("error when adding %q to config: %v", key, err)
	}

	err = cfg.Unmarshal(&C)
	if err != nil {
		return fmt.Errorf("invalid configuration after setting %q: %v", key, err)
	}

	if err = validateConfig(&C); err != nil {
		return fmt.Errorf("config validation failed after setting %q: %v", key, err)
	}

	return nil
}
