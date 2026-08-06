package runtimes

import (
	"errors"

	"github.com/AaltoRSE/shark-tank/internal/config"
	"github.com/AaltoRSE/shark-tank/internal/types"
)

type Runtime interface {
	Run(env types.SharkEnv, args []string) (int, error)
	Shell(env types.SharkEnv) (int, error)
}

func GetRuntime(name string) (Runtime, error) {

	runtimes := config.GetRuntimes()

	runtimeSpec, ok := runtimes[name].(types.RuntimeSpec)
	if !ok {
		return nil, errors.New("runtime not found")
	}

	// Call the right constructor based on the runtime type
	switch runtimeSpec.Type {
	case "apptainer":
		runtimeSpec, ok := runtimes[name].(types.ApptainerRuntimeSpec)
		if !ok {
			return nil, errors.New("runtime is not of type ApptainerRuntimeSpec")
		}
		imageUrl := runtimeSpec.ImageUrl
		cacheDir := runtimeSpec.CacheDir
		spec := NewApptainerRuntimeSpec(imageUrl, cacheDir)
		return NewApptainerRuntimeFromSpec(spec), nil
	}

	return nil, errors.New("no runtime found")
}
