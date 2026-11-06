// Package update implements the self-update of the moat binary: it
// checks the upstream GitHub repository for a newer release, asks the
// user whether the binary should be updated and whether the current
// binary should first be backed up, downloads the release archive for
// the current platform, verifies its SHA-256 checksum against the
// release checksums, and replaces the running binary with the extracted
// moat binary.
package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AaltoRSE/moat/internal/prompts"
	"github.com/AaltoRSE/moat/internal/version"
)

// binaryName is the name of the moat binary: both the name of the binary
// inside the release archives and the base name used for the backup copy
// of the current binary.
const binaryName = "moat"

// Variables that tests replace with test doubles.
var (
	// executablePath returns the path of the running moat binary.
	executablePath = os.Executable

	// askUpdate asks the user whether the binary should be updated to
	// the given version.
	askUpdate = prompts.ConfirmUpdate

	// askBackup asks the user whether the current binary should be
	// backed up under the given file name.
	askBackup = prompts.ConfirmBackup
)

// Update checks the upstream repository for a newer moat release than
// the version of the running binary. When a newer release is available,
// it asks the user whether the binary should be updated and whether the
// current binary should first be backed up as moat-<current version> in
// the directory of the running binary. It then downloads the release
// archive for the current operating system and architecture, verifies its
// SHA-256 checksum against the release checksums, extracts the moat
// binary, and replaces the running binary with it. It returns an error
// when the release information cannot be fetched, the download or the
// checksum verification fails, or the binary cannot be replaced. No error
// is returned when the binary is already up to date or when the user
// declines the update.
func Update() error {
	fmt.Println("Checking for the latest moat release...")

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("fetching the latest release failed: %w", err)
	}

	// Parse the versions immediately: a release tag that is not a
	// semantic version cannot be compared and is an error, while the
	// current version cannot be parsed for a binary built without
	// version information, in which case the user is asked whether to
	// update anyway.
	latest, ok := version.CreateSemVersion(release.TagName)
	if !ok {
		return fmt.Errorf("the latest release tag %q is not a semantic version", release.TagName)
	}
	current, currentKnown := version.CreateSemVersion(version.MoatVersion)
	if !currentKnown {
		fmt.Println("Could not determine the current version of moat.")
	} else if version.CompareSemVersion(latest, current) <= 0 {
		fmt.Printf("moat is up to date (%s).\n", current.Raw)
		return nil
	}

	fmt.Printf("Current version: %s\n", displayVersion(current.Raw))
	fmt.Printf("Latest release: %s\n", latest.Raw)

	update, err := askUpdate(latest.Raw)
	if err != nil {
		return fmt.Errorf("the update prompt failed: %w", err)
	}
	if !update {
		fmt.Println("Update aborted.")
		return nil
	}

	binaryPath, err := currentBinaryPath()
	if err != nil {
		return err
	}

	backup := backupName(current.Raw)
	saveBackup, err := askBackup(backup)
	if err != nil {
		return fmt.Errorf("the backup prompt failed: %w", err)
	}
	backupPath := ""
	if saveBackup {
		backupPath = filepath.Join(filepath.Dir(binaryPath), backup)
	}

	platform := version.PlatformName()
	fmt.Printf("Downloading moat %s (%s)...\n", latest.Raw, platform)
	archivePath, checksumsPath, cleanup, err := downloadReleaseAssets(release, platform, latest.Raw)
	if err != nil {
		return fmt.Errorf("downloading the release assets failed: %w", err)
	}
	defer cleanup()

	fmt.Println("Verifying the checksum of the downloaded archive...")
	if err := verifyChecksum(archivePath, checksumsPath, archiveName(platform)); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	newBinaryPath, err := extractBinary(archivePath, filepath.Dir(binaryPath))
	if err != nil {
		return fmt.Errorf("extracting the moat binary failed: %w", err)
	}

	fmt.Printf("Updating the binary at %s...\n", binaryPath)
	if err := installBinary(binaryPath, newBinaryPath, backupPath); err != nil {
		return fmt.Errorf("replacing the binary failed: %w", err)
	}

	if backupPath != "" {
		fmt.Printf("Previous binary saved as %s.\n", backupPath)
	}
	fmt.Printf("moat updated to %s.\n", latest.Raw)
	return nil
}

// currentBinaryPath returns the path of the running moat binary with any
// symbolic links resolved, so that the updated binary and the backup are
// placed next to the actual binary file. It returns an error when the
// path of the running binary cannot be determined.
func currentBinaryPath() (string, error) {
	path, err := executablePath()
	if err != nil {
		return "", fmt.Errorf("determining the path of the running binary failed: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Abs(path)
}

// backupName returns the file name for the backup copy of the current
// binary: the binary name followed by the given version, for example
// moat-v0.3.3. The leading "v" of the version is kept. When the version
// is unknown (empty), the name moat-unknown is returned.
func backupName(currentVersion string) string {
	if currentVersion == "" {
		return binaryName + "-unknown"
	}
	return binaryName + "-" + currentVersion
}

// displayVersion returns the version string shown to the user: the given
// version, or "unknown" when the version is unknown (empty).
func displayVersion(v string) string {
	if v == "" {
		return "unknown"
	}
	return v
}
