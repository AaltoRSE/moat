package prompts

import (
	"fmt"
)

// ConfirmUpdate asks the user whether the moat binary should be updated
// to the given version (for example v0.3.4) and reports the answer.
func ConfirmUpdate(latest string) (bool, error) {
	return confirm(fmt.Sprintf("Do you want to update moat to %s?", latest))
}

// ConfirmBackup asks the user whether the current moat binary should be
// backed up under the given file name (for example moat-v0.3.3) before
// the update and reports the answer.
func ConfirmBackup(backupName string) (bool, error) {
	return confirm(fmt.Sprintf("Do you want to save a backup of the current binary as %s?", backupName))
}
