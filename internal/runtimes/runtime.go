package runtimes

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

type Runtime interface {
	Run(env types.MoatEnv, args []string, envVars []string) (int, error)
	Shell(env types.MoatEnv) (int, error)
}

func GetRuntime(cfg *viper.Viper, name string) (Runtime, error) {

	var (
		runtimeSpec types.RuntimeSpec
	)
	runtimeSpec, err := config.GetRuntimeSpec(cfg, name)
	if err != nil {
		return nil, err
	}

	// Call the right constructor based on the runtime type
	switch runtimeSpec.Type {
	case "apptainer":
		return NewApptainerRuntimeFromSpec(&runtimeSpec), nil
	}

	return nil, fmt.Errorf("no runtime found")
}
