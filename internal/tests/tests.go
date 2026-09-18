// Package tests provides shared helpers for building isolated, temporary
// moat configurations in tests.
package tests

import (
	"os"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/env"
	"github.com/AaltoRSE/moat/internal/logging"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
)

var MoatTestDir string = "/tmp/moat_tests"

// CreateTempConfig initializes a new temporary moat configuration, creates
// the given environments and runtimes in it, and returns the path to the
// temporary config file along with its current contents.
//
// The configuration is written to a new file in the system temporary
// directory. Missing directories referenced by the environments (the fake
// home directories and mount source paths) are created automatically so that
// the helper never blocks on an interactive prompt. The runtimes are stored
// under the top-level runtimes key, where they overwrite default runtimes
// with the same name. The caller is responsible for removing the temporary
// config file when it is no longer needed.
func CreateTempConfig(envs map[string]types.MoatEnv, runtimes map[string]types.RuntimeSpec) (string, string, error) {
	// Create Moat temporary directory
	err := os.MkdirAll(MoatTestDir, 0755)
	if err != nil {
		return "", "", err
	}

	// Create a temporary config file
	configFile, err := os.CreateTemp(MoatTestDir, "moat-config.*.yaml")
	if err != nil {
		return "", "", err
	}
	configPath := configFile.Name()
	if err := configFile.Close(); err != nil {
		return "", "", err
	}

	// Silence standard output and initialize logging during setup
	capture := utils.OutputCapture{}
	capture.StartCapture()
	defer func() {
		_, _ = capture.StopCapture()
	}()
	logging.InitLogging(false)

	// Initialize the configuration from the temporary file
	conf, err := config.InitConfig(configPath)
	if err != nil {
		return "", "", err
	}

	// Create the environments in the configuration
	for name, moatEnv := range envs {
		if err := env.CreateEnvironment(conf, name, moatEnv, true); err != nil {
			return "", "", err
		}
	}

	// Set the runtimes in the configuration
	if len(runtimes) > 0 {
		conf.Set("runtimes", runtimes)
		if err := config.WriteConfig(conf); err != nil {
			return "", "", err
		}
	}

	// Read back the config file contents
	contents, err := os.ReadFile(configPath)
	if err != nil {
		return "", "", err
	}

	return configPath, string(contents), nil
}
