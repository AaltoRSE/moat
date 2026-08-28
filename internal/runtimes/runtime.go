package runtimes

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

type Runtime interface {
	Run(env types.MoatEnv, args []string) (int, error)
	Shell(env types.MoatEnv) (int, error)
}

func GetRuntime(name string) (Runtime, error) {

	var (
		runtimeSpecMap map[string]any
	)

	runtimeSpecMap = viper.GetStringMap("runtimes." + name)
	if len(runtimeSpecMap) == 0 {
		runtimeSpecMap = viper.GetStringMap("defaults.runtimes." + name)
		log.Debug().Interface("runtimeSpecMap", runtimeSpecMap).Msg("Loaded runtime spec map")
		if len(runtimeSpecMap) == 0 {
			return nil, fmt.Errorf("runtime not found")
		}
	}

	// Call the right constructor based on the runtime type
	switch runtimeSpecMap["type"] {
	case "apptainer":
		spec := NewApptainerRuntimeSpec(
			runtimeSpecMap["imageurl"].(string),
			runtimeSpecMap["cachedir"].(string),
			runtimeSpecMap["passenv"].(bool),
		)
		log.Debug().Interface("spec", spec).Msg("Created Apptainer runtime spec")
		return NewApptainerRuntimeFromSpec(spec), nil
	}

	return nil, fmt.Errorf("no runtime found")
}
