package cmd

import (
	"os"
	"testing"

	"github.com/AaltoRSE/moat/cmd"
	"github.com/AaltoRSE/moat/internal/tests"
	types "github.com/AaltoRSE/moat/internal/types"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/assert/yaml"
	"github.com/stretchr/testify/suite"
)

type EnvTestSuite struct {
	suite.Suite
	ConfigFile     string
	ExpectedOutput string
	ExpectedEnv    types.MoatEnv
}

func (suite *EnvTestSuite) SetupTest() {
	// Add environment called test to a fresh temporary moat-config
	suite.ExpectedEnv = types.MoatEnv{Home: "/tmp", Mounts: []string{}, ReadOnlyMounts: []string{}, Command: ""}
	var err error
	suite.ConfigFile, suite.ExpectedOutput, err = tests.CreateTempConfig("test", suite.ExpectedEnv)
	if err != nil {
		panic(err)
	}
}

func (suite *EnvTestSuite) TearDownTest() {
	// Remove the temporary config file
	err := os.Remove(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
}

// TestShow tests that the env show command prints the configuration of the
// named environment.
func (suite *EnvTestSuite) TestShow() {

	capture := utils.OutputCapture{}
	capture.StartCapture()
	cmd.RootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "show", "--name", "test"})
	err := cmd.RootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}

	// Unmarshal output
	var outputEnv types.MoatEnv
	err = yaml.Unmarshal([]byte(capturedOutput), &outputEnv)
	if err != nil {
		panic(err)
	}
	assert.Equal(suite.T(), suite.ExpectedEnv, outputEnv)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEnvTestSuite(t *testing.T) {
	suite.Run(t, new(EnvTestSuite))
}
