package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
	ConfigFile string
	BaseOutput string
	BaseEnv    types.MoatEnv
}

func (suite *ConfigTestSuite) SetupTest() {
	// Add environment called test to a fresh temporary moat-config
	suite.BaseEnv = types.MoatEnv{Home: tests.MoatTestDir, Mounts: nil, ReadOnlyMounts: nil, Command: nil}
	var err error
	suite.ConfigFile, suite.BaseOutput, err = tests.CreateTempConfig("test", suite.BaseEnv)
	if err != nil {
		panic(err)
	}
}

func (suite *ConfigTestSuite) TearDownTest() {
	// Remove the temporary config file
	err := os.Remove(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
}

// TestGetConfigFile tests that GetConfigFile returns the active
// configuration file and falls back to the global moat-config.yaml when
// there is no active configuration.
func TestGetConfigFile(t *testing.T) {
	cfgFile, err := os.CreateTemp("", "moat-config.*.yaml")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Remove(cfgFile.Name()))
	}()
	err = cfgFile.Close()
	assert.NoError(t, err)

	cfg, err := config.InitConfig(cfgFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, cfgFile.Name(), config.GetConfigFile(cfg))

	home, err := os.UserHomeDir()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "moat", "moat-config.yaml"), config.GetConfigFile(nil))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
