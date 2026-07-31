package config

import (
	"errors"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	yaml "go.yaml.in/yaml/v3"
)

type Config struct {
	Defaults Defaults    `validate:"required"`
	Envs     []EnvConfig `validate:""`
}

type ApptainerInstanceRuntime struct {
	ImageUrl string `validate:"required"`
	CacheDir string `validate:"required,filepath"`
}

type Defaults struct {
	Runtime       string                   `validate:"required"`
	RuntimeConfig ApptainerInstanceRuntime `validate:"required_if=Runtime apptainerinstance"`
}

type EnvConfig struct {
	Name   string   `validate:"required"`
	Home   string   `validate:"required"`
	Mounts []string `validate:"required,dive,required"`
}

func InitConfig() {
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
	viper.ReadInConfig()

	var C Config
	err := viper.Unmarshal(&C)
	if err != nil {
		log.Fatalf("unable to unmarshal config: %v", err)
	}

	if err := validateConfig(&C); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
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

func validateConfig(config *Config) error {
	// Validate that the configuration struct is properly filled out
	var validate *validator.Validate
	validate = validator.New(validator.WithRequiredStructEnabled())
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
