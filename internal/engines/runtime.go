// Package engines provides the runtime engines that execute commands and
// shells inside container environments.
//
// Each engine implements the [Runtime] interface, which hides the
// container technology from the callers. The [GetRuntime] function builds
// the engine named in the configuration from its runtime spec.
package engines

import (
	"fmt"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/types"
	"github.com/spf13/viper"
)

// Runtime is the interface implemented by all runtime engines. A runtime
// engine executes a command or opens a shell inside the container
// environment described by a [types.MoatEnv].
type Runtime interface {
	Run(env types.MoatEnv, args []string, envVars []string) (int, error)
	Shell(env types.MoatEnv) (int, error)
	CacheList() ([]string, error)
}

// GetRuntime returns the runtime engine named name, built from the
// runtime spec looked up in the configuration. It returns an error if the
// runtime spec is not found or the runtime type is unknown.
func GetRuntime(cfg *viper.Viper, name string) (Runtime, error) {

	var (
		runtimeSpec types.RuntimeSpec
	)
	runtimeSpec, err := config.GetRuntimeSpec(cfg, name)
	if err != nil {
		return nil, err
	}

	// Safety net for hand-edited configurations: overlay settings are only
	// supported by the apptainer runtime (config validation normally
	// rejects them for other types already).
	if runtimeSpec.Type != "apptainer" &&
		(runtimeSpec.OverlayOwner != "" || runtimeSpec.OverlayDriver != "" || runtimeSpec.OverlayFSSnapshot != "" || runtimeSpec.OverlayDir != "") {
		return nil, fmt.Errorf("runtime %q: overlay settings are only supported for the apptainer type", name)
	}

	// Call the right constructor based on the runtime type
	switch runtimeSpec.Type {
	case "apptainer":
		return NewApptainerRuntimeFromSpec(&runtimeSpec), nil
	case "docker":
		return NewDockerRuntimeFromSpec(&runtimeSpec), nil
	}

	return nil, fmt.Errorf("no runtime found")
}
