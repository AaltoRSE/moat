package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	yaml "go.yaml.in/yaml/v3"
)

func InitConfig() error {
	viper.SetConfigName("config")

	// Set defaults if not set
	viper.SetDefault("Defaults",
		map[string]any{
			"runtime": "apptainerinstance",
			"runtimeconfig": map[string]string{
				"imageUrl": "ghcr.io/aaltorse/vscode-apptainer:latest",
				"cachedir": "$HOME/.cache/shark-tank/images",
			},
		},
	)

	viper.AddConfigPath("$HOME/.shark-tank")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error reading config file: %v", err)
		}
	}

	var C types.Config
	err := viper.Unmarshal(&C)
	if err != nil {
		fmt.Println("Unable to unmarshal config", err)
		// Print the configuration as a string for debugging
		configStr := GetConfigAsString()
		log.Printf("Current configuration:\n%s", configStr)
		return fmt.Errorf("unable to unmarshal config: %v", err)
	}

	if err := validateConfig(&C); err != nil {
		configStr := GetConfigAsString()
		log.Printf("Current configuration:\n%s", configStr)
		return fmt.Errorf("config validation failed: %v", err)
	}

	return nil
}

// WriteConfig writes the current viper configuration to the config file.
func WriteConfig() error {
	err := viper.WriteConfig()
	if err != nil {
		log.Printf("Error writing config: %v", err)
		return err
	}
	return nil
}

func validateConfig(config *types.Config) error {
	// Validate that the configuration struct is properly filled out
	var validate = validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(config)
	if err != nil {

		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			for _, e := range validateErrs {
				log.Print(e.Namespace())
				log.Print(e.Field())
				log.Print(e.StructNamespace())
				log.Print(e.StructField())
				log.Print(e.Tag())
				log.Print(e.ActualTag())
				log.Print(e.Kind())
				log.Print(e.Type())
				log.Print(e.Value())
				log.Print(e.Param())
				log.Print()
			}
		}
		return err
	}
	return nil
}

func GetConfigAsString() string {
	c := viper.AllSettings()
	bs, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("unable to marshal config to YAML: %v", err)
	}
	return string(bs)
}

func GetEnv(name string) (types.SharkEnv, error) {
	envMap := viper.GetStringMap("envs." + name)
	if len(envMap) == 0 {
		return types.SharkEnv{}, fmt.Errorf("environment %q not found", name)
	}
	home, _ := envMap["home"].(string)
	var mounts []string
	if raw, ok := envMap["mounts"].([]interface{}); ok {
		for _, v := range raw {
			if s, ok := v.(string); ok {
				mounts = append(mounts, s)
			}
		}
	}
	return types.SharkEnv{
		Name:     name,
		FakeHome: home,
		Mounts:   mounts,
	}, nil
}
