package update

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// installBinary replaces the file at binaryPath with the file at
// newBinaryPath. The previous binary is moved to backupPath when
// backupPath is not empty; an existing file at backupPath is overwritten.
// Without a backupPath the previous binary is moved aside and deleted.
// Moving the previous binary away first allows replacing a binary that is
// currently executing. It returns an error when the replacement fails;
// the previous binary is restored to binaryPath in that case.
func installBinary(binaryPath, newBinaryPath, backupPath string) error {
	oldPath := binaryPath + ".old"
	if backupPath != "" {
		oldPath = backupPath
	}
	if err := os.Rename(binaryPath, oldPath); err != nil {
		return fmt.Errorf("moving the previous binary out of the way failed: %w", err)
	}
	if err := os.Rename(newBinaryPath, binaryPath); err != nil {
		if restoreErr := os.Rename(oldPath, binaryPath); restoreErr != nil {
			return fmt.Errorf("%v (restoring the previous binary failed: %v)", err, restoreErr)
		}
		return err
	}
	if backupPath == "" {
		if err := os.Remove(oldPath); err != nil {
			log.Warn().Err(err).Str("path", oldPath).Msg("Failed to remove the previous binary")
		}
	}
	return nil
}
