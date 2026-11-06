package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// verifyChecksum compares the SHA-256 checksum of the file at archivePath
// with the checksum recorded for fileName in the checksums file at
// checksumsPath. It returns an error when the checksums file contains no
// entry for fileName or the checksums do not match.
func verifyChecksum(archivePath, checksumsPath, fileName string) error {
	expected, err := expectedChecksum(checksumsPath, fileName)
	if err != nil {
		return err
	}
	actual, err := fileChecksum(archivePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", fileName, expected, actual)
	}
	return nil
}

// expectedChecksum returns the SHA-256 checksum recorded for fileName in
// the checksums file at path. Each line of the file has the sha256sum
// format "<checksum>  <filename>"; lines that are empty or do not match
// this format are ignored. It returns an error when the file cannot be
// read or contains no entry for fileName.
func expectedChecksum(path, fileName string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading the checksums file failed: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || !isSHA256Hex(fields[0]) {
			continue
		}
		if fields[1] == fileName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("checksums file contains no checksum for %s", fileName)
}

// fileChecksum returns the hex-encoded SHA-256 checksum of the file at
// path.
func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// isSHA256Hex reports whether s is a hexadecimal string of the length of
// a SHA-256 checksum.
func isSHA256Hex(s string) bool {
	if len(s) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
