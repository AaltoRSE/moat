package prompts_test

import (
	"testing"

	"github.com/AaltoRSE/moat/internal/prompts"
	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UpdatePromptTestSuite tests the confirmation prompts of the moat
// update: whether the binary should be updated and whether the current
// binary should be backed up.
type UpdatePromptTestSuite struct {
	suite.Suite
}

// TestConfirmUpdateAccept tests that the update prompt reports true
// when the user accepts.
func (suite *UpdatePromptTestSuite) TestConfirmUpdateAccept() {
	tests.AnswerPrompt(suite.T(), "y")

	update, err := prompts.ConfirmUpdate("v0.3.4")

	require.NoError(suite.T(), err)
	assert.True(suite.T(), update)
}

// TestConfirmUpdateDecline tests that the update prompt reports false
// when the user declines.
func (suite *UpdatePromptTestSuite) TestConfirmUpdateDecline() {
	tests.AnswerPrompt(suite.T(), "n")

	update, err := prompts.ConfirmUpdate("v0.3.4")

	require.NoError(suite.T(), err)
	assert.False(suite.T(), update)
}

// TestConfirmBackupAccept tests that the backup prompt reports true when
// the user accepts.
func (suite *UpdatePromptTestSuite) TestConfirmBackupAccept() {
	tests.AnswerPrompt(suite.T(), "y")

	backup, err := prompts.ConfirmBackup("moat-v0.3.3")

	require.NoError(suite.T(), err)
	assert.True(suite.T(), backup)
}

// TestConfirmBackupDecline tests that the backup prompt reports false
// when the user declines.
func (suite *UpdatePromptTestSuite) TestConfirmBackupDecline() {
	tests.AnswerPrompt(suite.T(), "n")

	backup, err := prompts.ConfirmBackup("moat-v0.3.3")

	require.NoError(suite.T(), err)
	assert.False(suite.T(), backup)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUpdatePromptTestSuite(t *testing.T) {
	suite.Run(t, new(UpdatePromptTestSuite))
}
