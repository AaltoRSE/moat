package env_test

import (
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/cmd/root"
	"github.com/AaltoRSE/moat/internal/config"
	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/stretchr/testify/assert"
)

// TestCopyWithRoMounts tests that env copy overrides the source environment's
// read-only mounts when the --ro-mount flag is provided.
func (suite *EnvTestSuite) TestCopyWithRoMounts() {
	baseDir := filepath.Join(tests.MoatTestDir, "copy_ro_mounts")
	home := filepath.Join(baseDir, "home")
	roMount := filepath.Join(baseDir, "romount")
	// Start from a clean slate so the test verifies directory creation
	err := os.RemoveAll(baseDir)
	if err != nil {
		panic(err)
	}
	suite.T().Cleanup(func() {
		_ = os.RemoveAll(baseDir)
	})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	rootCmd := root.CreateRootCmd()
	rootCmd.SetArgs([]string{"--config", suite.ConfigFile, "env", "copy", "--source", "test", "--name", "newenv", "--home", home, "--ro-mount", roMount, "--yes"})
	err = rootCmd.Execute()
	if err != nil {
		panic(err)
	}
	capturedOutput, err := capture.StopCapture()
	if err != nil {
		panic(err)
	}
	assert.Contains(suite.T(), capturedOutput, "Environment created successfully.")

	// The home and read-only mount source directories must have been created
	assert.DirExists(suite.T(), home)
	assert.DirExists(suite.T(), roMount)

	// Unmarshal the created configuration and check the stored environment
	cfg, err := config.InitConfig(suite.ConfigFile)
	if err != nil {
		panic(err)
	}
	envs := config.GetEnvs(cfg)
	createdEnv, exists := envs["newenv"]
	assert.True(suite.T(), exists)
	assert.Equal(suite.T(), home, createdEnv.Home)
	assert.Equal(suite.T(), []string{roMount}, createdEnv.ReadOnlyMounts)
}
