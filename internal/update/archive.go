package update

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/AaltoRSE/moat/internal/utils"
)

// extractBinary extracts the moat binary from the gzipped tar archive at
// archivePath into a new file created in destDir and returns its path.
// Only the archive entry named "moat" is extracted; the other entries
// (e.g. the license and the readme) are skipped. The extracted file is
// made executable. It returns an error when the archive cannot be read or
// does not contain the moat binary as a regular file.
func extractBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer func() { _ = gz.Close() }()

	tarReader := tar.NewReader(gz)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return "", fmt.Errorf("archive does not contain the moat binary")
		}
		if err != nil {
			return "", err
		}
		// Match only the moat binary at the root of the archive; this
		// also rules out path traversal in entry names.
		if path.Clean(header.Name) != binaryName {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() {
			return "", fmt.Errorf("archive entry %q is not a regular file", header.Name)
		}
		return writeBinary(tarReader, destDir)
	}
}

// writeBinary writes the binary read from r to a new executable file in
// destDir and returns its path. The file is removed when writing it
// fails.
func writeBinary(r io.Reader, destDir string) (string, error) {
	out, err := os.CreateTemp(destDir, "."+binaryName+".update-*")
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, r); err != nil {
		_ = out.Close()
		_ = utils.CleanupFiles([]string{out.Name()})
		return "", err
	}
	if err := out.Chmod(0o755); err != nil {
		_ = out.Close()
		_ = utils.CleanupFiles([]string{out.Name()})
		return "", err
	}
	if err := out.Close(); err != nil {
		_ = utils.CleanupFiles([]string{out.Name()})
		return "", err
	}
	return out.Name(), nil
}
