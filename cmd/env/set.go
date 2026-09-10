package env

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/cobra"
)

// CreateEnvSetCmd creates the set subcommand for EnvCmd.
func CreateEnvSetCmd() *cobra.Command {

	// setCmd is the cobra command that sets a variable of an existing
	// environment.
	var setCmd = &cobra.Command{
		Use:   "set <key> <value> [<value>]",
		Short: "Set a variable of an existing moat environment",
		Long: `Set a variable of an existing moat environment.

This command allows you to set a specific variable of an existing
environment. Only the environment given by the -n/--name flag is
modified; all other environments and the rest of the configuration
are left untouched.

Examples:
  moat env set -n test command "code --wait ."
  moat env set -n test mountcwd true
  moat env set -n test mounts /path/to/project /run/dbus`,
		Args: cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Fatal().Msgf("could not get name flag: %v", err)
			}

			// The environment must already exist
			if _, err := config.GetEnv(config.CmdConfig, name, false); err != nil {
				log.Fatal().Msgf("failed to get environment %q: %v", name, err)
			}

			key := args[0]
			fullKey := "envs." + name + "." + key

			// Check key type; fall back to the MoatEnv struct definition
			// for optional fields that are not set yet
			keyType, err := config.GetVariableType(config.CmdConfig, fullKey)
			if err != nil {
				keyType, err = moatEnvFieldType(key)
				if err != nil {
					log.Fatal().Msgf("failed to get type for key %q: %v", key, err)
				}
			}
			if keyType.Kind() == reflect.Pointer {
				keyType = keyType.Elem()
			}

			var value any

			switch keyType.Kind() {
			case reflect.String:
				// Return error if there are more than 2 arguments for a string key
				if len(args) > 2 {
					log.Fatal().Msgf("too many arguments for key %q of type string", key)
				}
				value = args[1]
			case reflect.Slice:
				value = args[1:]
			case reflect.Bool:
				if len(args) > 2 {
					log.Fatal().Msgf("too many arguments for key %q of type bool", key)
				}
				value, err = strconv.ParseBool(args[1])
				if err != nil {
					log.Fatal().Msgf("invalid boolean value %q for key %q: %v", args[1], key, err)
				}
			default:
				log.Fatal().Msgf("unsupported key type for key %q: %v", key, keyType.Kind())
			}

			if err := config.SetConfig(config.CmdConfig, fullKey, value); err != nil {
				log.Fatal().Msgf("failed to set %q: %v", fullKey, err)
			}
			fmt.Printf("Set %s.%s = %v\n", name, key, value)
		},
	}

	setCmd.Flags().StringP("name", "n", "", "Name of the environment to modify")

	if err := setCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}

	return setCmd
}

// moatEnvFieldType returns the reflect.Type of the MoatEnv field whose YAML
// tag (or field name) matches key. It returns an error if no such field
// exists.
func moatEnvFieldType(key string) (reflect.Type, error) {
	envType := reflect.TypeOf(types.MoatEnv{})
	for i := 0; i < envType.NumField(); i++ {
		field := envType.Field(i)
		tag := strings.Split(field.Tag.Get("yaml"), ",")[0]
		if tag == key || field.Name == key {
			return field.Type, nil
		}
	}
	return nil, fmt.Errorf("key %q is not a valid environment variable", key)
}
