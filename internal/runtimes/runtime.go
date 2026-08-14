package runtimes

import (
	"errors"
	"fmt"

	"github.com/AaltoRSE/shark-tank/internal/types"
	"github.com/spf13/viper"
)

type Runtime interface {
	Run(env types.SharkEnv, args []string) (int, error)
	Shell(env types.SharkEnv) (int, error)
}

func GetRuntime(name string) (Runtime, error) {

	////runtimes := config.GetRuntimes()

	var (
		runtimeSpecMap map[string]string
	)

	runtimeSpecMap = viper.GetStringMapString("runtimes." + name)
	if len(runtimeSpecMap) == 0 {
		runtimeSpecMap = viper.GetStringMapString("defaults.runtimes." + name)
		fmt.Printf("runtimeSpecMap: %v\n", runtimeSpecMap)
		if len(runtimeSpecMap) == 0 {
			return nil, errors.New("runtime not found")
		}
	}

	//runtimeSpec, ok := runtimes[name].(types.RuntimeSpec)
	//if !ok {
	//	return nil, errors.New("runtime not found")
	//}

	// Call the right constructor based on the runtime type
	switch runtimeSpecMap["type"] {
	case "apptainer":
		spec := NewApptainerRuntimeSpec(
			runtimeSpecMap["imageurl"],
			runtimeSpecMap["cachedir"],
		)
		fmt.Printf("Created Apptainer runtime spec: %+v\n", spec)
		return NewApptainerRuntimeFromSpec(spec), nil
	}

	return nil, errors.New("no runtime found")
}
