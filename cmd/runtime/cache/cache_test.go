package cache_test

import (
	"os"
	"testing"

	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CacheTestSuite is the shared test suite for the `moat runtime cache`
// command group. Each test starts from a fresh temporary moat
// configuration that contains no environments and no user-specified
// runtimes, so that only the default runtimes are available.
type CacheTestSuite struct {
	suite.Suite
	ConfigFile string
}

func (suite *CacheTestSuite) SetupTest() {
	// Use a fresh temporary moat-config without any extra environments or
	// user-specified runtimes
	var err error
	suite.ConfigFile, _, err = tests.CreateTempConfig(nil, nil)
	require.NoError(suite.T(), err)
}

func (suite *CacheTestSuite) TearDownTest() {
	// Remove the temporary config file
	suite.Require().NoError(os.Remove(suite.ConfigFile))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}
