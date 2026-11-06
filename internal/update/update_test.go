package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaltoRSE/moat/internal/tests"
	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/AaltoRSE/moat/internal/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// UpdateTestSuite tests the self-update of the moat binary: the
// version comparison, the download and checksum verification of the
// release assets, the extraction and installation of the new binary, and
// the full update flow against a fake release server.
type UpdateTestSuite struct {
	suite.Suite
	// products holds the paths of the files the current test has
	// created, for the cleanup registered by newTestDir.
	products []string
}

// SetupTest creates the base directory for the test files.
func (suite *UpdateTestSuite) SetupTest() {
	require.NoError(suite.T(), os.MkdirAll(tests.MoatTestDir, 0755))
}

// newTestDir creates a fresh directory for the test and registers the
// removal of the test's products and of the directory itself when the
// test finishes: the recorded products are removed with
// utils.CleanupFiles and the directory is then removed with
// utils.CleanupDir, which fails when the directory still contains
// unknown files.
func (suite *UpdateTestSuite) newTestDir() string {
	dir := filepath.Join(tests.MoatTestDir, "update", suite.T().Name())
	require.NoError(suite.T(), utils.CleanupDir(dir))
	require.NoError(suite.T(), os.MkdirAll(dir, 0755))
	suite.T().Cleanup(func() {
		require.NoError(suite.T(), utils.CleanupFiles(suite.products))
		require.NoError(suite.T(), utils.CleanupDir(dir))
	})
	return dir
}

// track records the paths of files created by the current test so that
// the cleanup registered by newTestDir removes them.
func (suite *UpdateTestSuite) track(paths ...string) {
	suite.products = append(suite.products, paths...)
}

// setMoatSource points version.MoatSource at the given endpoint and
// restores the original value when the test finishes.
func (suite *UpdateTestSuite) setMoatSource(url string) {
	original := version.MoatSource
	suite.T().Cleanup(func() { version.MoatSource = original })
	version.MoatSource = url
}

// setExecutablePath points executablePath at the given path and restores
// the original value when the test finishes.
func (suite *UpdateTestSuite) setExecutablePath(path string) {
	original := executablePath
	suite.T().Cleanup(func() { executablePath = original })
	executablePath = func() (string, error) { return path, nil }
}

// setVersion sets the build-time version metadata to the given version
// and restores the original value when the test finishes.
func (suite *UpdateTestSuite) setVersion(v string) {
	original := version.MoatVersion
	suite.T().Cleanup(func() { version.MoatVersion = original })
	version.MoatVersion = v
}

// promptAnswers holds the answers given by the stubbed update prompts
// and records whether the prompts were called at all.
type promptAnswers struct {
	updateAnswer bool
	backupAnswer bool
	updateCalled bool
	backupCalled bool
}

// setPromptAnswers replaces the update prompts with stubs that return
// the given answers and record the calls in answers. The original
// prompts are restored when the test finishes.
func (suite *UpdateTestSuite) setPromptAnswers(answers *promptAnswers) {
	originalAskUpdate, originalAskBackup := askUpdate, askBackup
	suite.T().Cleanup(func() {
		askUpdate, askBackup = originalAskUpdate, originalAskBackup
	})
	askUpdate = func(string) (bool, error) {
		answers.updateCalled = true
		return answers.updateAnswer, nil
	}
	askBackup = func(string) (bool, error) {
		answers.backupCalled = true
		return answers.backupAnswer, nil
	}
}

// buildArchive creates a gzipped tar archive containing a moat binary
// with the given content and a readme file, and returns the archive
// contents.
func buildArchive(binaryContent []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	if err := tarWriter.WriteHeader(&tar.Header{Name: binaryName, Mode: 0o755, Size: int64(len(binaryContent))}); err != nil {
		return nil, err
	}
	if _, err := tarWriter.Write(binaryContent); err != nil {
		return nil, err
	}
	readme := []byte("# moat\n")
	if err := tarWriter.WriteHeader(&tar.Header{Name: "README.md", Mode: 0o644, Size: int64(len(readme))}); err != nil {
		return nil, err
	}
	if _, err := tarWriter.Write(readme); err != nil {
		return nil, err
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// newReleaseServer starts a test server that mimics the GitHub API and
// the release downloads for a release with the given tag: the endpoint
// /repos/AaltoRSE/moat/releases/latest serves the release information,
// and /download/{name} serves the archive (built from binaryContent) and
// the checksums file. When corruptChecksums is true, the checksums file
// records a wrong checksum for the archive. The moat source that
// [getAPIUrl] maps to the releases API of the test server is returned.
func (suite *UpdateTestSuite) newReleaseServer(tag string, binaryContent []byte, corruptChecksums bool) string {
	platform := version.PlatformName()
	archive, err := buildArchive(binaryContent)
	require.NoError(suite.T(), err)
	archiveChecksum := fmt.Sprintf("%x", sha256.Sum256(archive))
	if corruptChecksums {
		archiveChecksum = fmt.Sprintf("%x", sha256.Sum256([]byte("not-the-archive")))
	}
	checksums := fmt.Sprintf("%s  %s\n", archiveChecksum, archiveName(platform))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/AaltoRSE/moat/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		downloadBase := "http://" + r.Host + "/download/"
		release := Release{
			TagName: tag,
			Assets: []Asset{
				{Name: archiveName(platform), BrowserDownloadURL: downloadBase + archiveName(platform)},
				{Name: checksumsName(tag), BrowserDownloadURL: downloadBase + checksumsName(tag)},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	})
	mux.HandleFunc("GET /download/{name}", func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("name") {
		case archiveName(platform):
			_, _ = w.Write(archive)
		case checksumsName(tag):
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(mux)
	suite.T().Cleanup(server.Close)
	return suite.testSource(server)
}

// testSource returns a moat source that [getAPIUrl] maps to the releases
// API endpoint of the given test server: the URL of the server followed
// by an owner and a repository name. The explicit http:// scheme of the
// test server URL is kept by [getAPIUrl] so that the plain test server
// can be addressed.
func (suite *UpdateTestSuite) testSource(server *httptest.Server) string {
	return server.URL + "/AaltoRSE/moat"
}

// dirEntries returns the sorted file names in dir.
func (suite *UpdateTestSuite) dirEntries(dir string) []string {
	dirEntries, err := os.ReadDir(dir)
	require.NoError(suite.T(), err)
	names := make([]string, 0, len(dirEntries))
	for _, entry := range dirEntries {
		names = append(names, entry.Name())
	}
	return names
}

// writeFakeBinary writes a fake moat binary with the given content to
// path with executable permissions and records the path for cleanup.
func (suite *UpdateTestSuite) writeFakeBinary(path string, content []byte) {
	suite.track(path)
	require.NoError(suite.T(), os.WriteFile(path, content, 0o755))
}

// readFakeBinary returns the content of the file at path.
func (suite *UpdateTestSuite) readFakeBinary(path string) []byte {
	content, err := os.ReadFile(path)
	require.NoError(suite.T(), err)
	return content
}

// runUpdate runs the update with the standard output captured and
// returns the captured output.
func (suite *UpdateTestSuite) runUpdate() string {
	capture := utils.OutputCapture{}
	capture.StartCapture()
	err := Update()
	output, captureErr := capture.StopCapture()
	require.NoError(suite.T(), captureErr)
	require.NoError(suite.T(), err)
	return output
}

// TestUpdateCurrentNewer tests that no prompt is asked when the current
// version is newer than the latest release, e.g. for a development build
// based on unreleased changes.
func (suite *UpdateTestSuite) TestUpdateCurrentNewer() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("current binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v1.0.0-5-gb22ebb1")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), false))
	answers := &promptAnswers{updateAnswer: true, backupAnswer: true}
	suite.setPromptAnswers(answers)

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "moat is up to date (v1.0.0-5-gb22ebb1)")
	assert.False(suite.T(), answers.updateCalled, "the update prompt should not be asked")
	assert.Equal(suite.T(), []byte("current binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestBackupName tests the file names of the backup copy of the current
// binary.
func (suite *UpdateTestSuite) TestBackupName() {
	assert.Equal(suite.T(), "moat-v0.3.3", backupName("v0.3.3"))
	assert.Equal(suite.T(), "moat-v0.3.3-14-gb22ebb1", backupName("v0.3.3-14-gb22ebb1"))
	assert.Equal(suite.T(), "moat-unknown", backupName(""))
}

// TestDisplayVersion tests the version strings shown to the user.
func (suite *UpdateTestSuite) TestDisplayVersion() {
	assert.Equal(suite.T(), "v0.3.3", displayVersion("v0.3.3"))
	assert.Equal(suite.T(), "unknown", displayVersion(""))
}

// TestArchiveAndChecksumsNames tests the release asset file names.
func (suite *UpdateTestSuite) TestArchiveAndChecksumsNames() {
	assert.Equal(suite.T(), "moat_Linux_x86_64.tar.gz", archiveName("Linux_x86_64"))
	assert.Equal(suite.T(), "moat_0.3.3_checksums.txt", checksumsName("v0.3.3"))
	assert.Equal(suite.T(), "moat_1.2.3_checksums.txt", checksumsName("v1.2.3"))
}

// TestGetAPIUrl tests the derivation of the GitHub API endpoint from a
// source repository.
func (suite *UpdateTestSuite) TestGetAPIUrl() {
	cases := []struct {
		source   string
		expected string
		ok       bool
	}{
		{"github.com/AaltoRSE/moat", "https://api.github.com/repos/AaltoRSE/moat/releases/latest", true},
		{"https://github.com/AaltoRSE/moat", "https://api.github.com/repos/AaltoRSE/moat/releases/latest", true},
		{"https://github.com/AaltoRSE/moat.git", "https://api.github.com/repos/AaltoRSE/moat/releases/latest", true},
		{"github.com/AaltoRSE/moat/", "https://api.github.com/repos/AaltoRSE/moat/releases/latest", true},
		{"AaltoRSE/moat", "https://api.github.com/repos/AaltoRSE/moat/releases/latest", true},
		{"http://127.0.0.1:1234/AaltoRSE/moat", "http://127.0.0.1:1234/repos/AaltoRSE/moat/releases/latest", true},
		{"github.example.com/AaltoRSE/moat", "https://github.example.com/repos/AaltoRSE/moat/releases/latest", true},
		{"moat", "", false},
		{"github.com/AaltoRSE/moat/extra/deep", "", false},
		{"", "", false},
	}
	for _, tc := range cases {
		apiURL, err := getAPIUrl(tc.source)
		if tc.ok {
			require.NoError(suite.T(), err, "unexpected error for source %q", tc.source)
			assert.Equal(suite.T(), tc.expected, apiURL, "unexpected API URL for source %q", tc.source)
		} else {
			require.Error(suite.T(), err, "expected an error for source %q", tc.source)
		}
	}
}

// TestFetchLatestRelease tests that the latest release is parsed from the
// GitHub API response.
func (suite *UpdateTestSuite) TestFetchLatestRelease() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": "v0.3.4", "assets": [{"name": "moat_Linux_x86_64.tar.gz", "browser_download_url": "https://example.com/moat_Linux_x86_64.tar.gz"}]}`))
	}))
	suite.T().Cleanup(server.Close)
	suite.setMoatSource(suite.testSource(server))

	release, err := fetchLatestRelease()

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "v0.3.4", release.TagName)
	require.Len(suite.T(), release.Assets, 1)
	assert.Equal(suite.T(), "moat_Linux_x86_64.tar.gz", release.Assets[0].Name)
}

// TestFetchLatestReleaseServerError tests that an error is returned when
// the API request fails.
func (suite *UpdateTestSuite) TestFetchLatestReleaseServerError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	suite.T().Cleanup(server.Close)
	suite.setMoatSource(suite.testSource(server))

	_, err := fetchLatestRelease()

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "500")
}

// TestFetchLatestReleaseInvalidJSON tests that an error is returned when
// the API response cannot be parsed.
func (suite *UpdateTestSuite) TestFetchLatestReleaseInvalidJSON() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not json"))
	}))
	suite.T().Cleanup(server.Close)
	suite.setMoatSource(suite.testSource(server))

	_, err := fetchLatestRelease()

	require.Error(suite.T(), err)
}

// TestFetchLatestReleaseNoTagName tests that an error is returned when
// the API response contains no tag name.
func (suite *UpdateTestSuite) TestFetchLatestReleaseNoTagName() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name": ""}`))
	}))
	suite.T().Cleanup(server.Close)
	suite.setMoatSource(suite.testSource(server))

	_, err := fetchLatestRelease()

	require.Error(suite.T(), err)
}

// TestDownloadFile tests that a file is downloaded to the given path.
func (suite *UpdateTestSuite) TestDownloadFile() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("downloaded content"))
	}))
	suite.T().Cleanup(server.Close)
	dest := filepath.Join(suite.newTestDir(), "downloaded")
	suite.track(dest)

	require.NoError(suite.T(), downloadFile(server.URL, dest))

	assert.Equal(suite.T(), []byte("downloaded content"), suite.readFakeBinary(dest))
}

// TestDownloadFileNotFound tests that an error is returned when the
// download fails.
func (suite *UpdateTestSuite) TestDownloadFileNotFound() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	suite.T().Cleanup(server.Close)
	dest := filepath.Join(suite.newTestDir(), "downloaded")

	err := downloadFile(server.URL, dest)

	require.Error(suite.T(), err)
	assert.NoFileExists(suite.T(), dest)
}

// TestVerifyChecksum tests the checksum verification of the downloaded
// archive.
func (suite *UpdateTestSuite) TestVerifyChecksum() {
	dir := suite.newTestDir()
	archivePath := filepath.Join(dir, "moat_Linux_x86_64.tar.gz")
	checksumsPath := filepath.Join(dir, "moat_0.3.4_checksums.txt")
	suite.writeFakeBinary(archivePath, []byte("archive content"))
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte("archive content")))
	checksums := fmt.Sprintf("garbage line\n%s  moat_Linux_x86_64.tar.gz\n%s  moat_Darwin_arm64.tar.gz\n",
		checksum, fmt.Sprintf("%x", sha256.Sum256([]byte("other"))))
	suite.track(checksumsPath)
	require.NoError(suite.T(), os.WriteFile(checksumsPath, []byte(checksums), 0o644))

	assert.NoError(suite.T(), verifyChecksum(archivePath, checksumsPath, "moat_Linux_x86_64.tar.gz"))
}

// TestVerifyChecksumMismatch tests that a checksum mismatch is reported.
func (suite *UpdateTestSuite) TestVerifyChecksumMismatch() {
	dir := suite.newTestDir()
	archivePath := filepath.Join(dir, "moat_Linux_x86_64.tar.gz")
	checksumsPath := filepath.Join(dir, "moat_0.3.4_checksums.txt")
	suite.writeFakeBinary(archivePath, []byte("archive content"))
	suite.track(checksumsPath)
	require.NoError(suite.T(), os.WriteFile(checksumsPath, []byte(fmt.Sprintf("%x  moat_Linux_x86_64.tar.gz\n", sha256.Sum256([]byte("different")))), 0o644))

	err := verifyChecksum(archivePath, checksumsPath, "moat_Linux_x86_64.tar.gz")

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "checksum mismatch")
}

// TestVerifyChecksumMissingEntry tests that a missing checksum entry is
// reported.
func (suite *UpdateTestSuite) TestVerifyChecksumMissingEntry() {
	dir := suite.newTestDir()
	archivePath := filepath.Join(dir, "moat_Linux_x86_64.tar.gz")
	checksumsPath := filepath.Join(dir, "moat_0.3.4_checksums.txt")
	suite.writeFakeBinary(archivePath, []byte("archive content"))
	suite.track(checksumsPath)
	require.NoError(suite.T(), os.WriteFile(checksumsPath, []byte("no checksums here\n"), 0o644))

	err := verifyChecksum(archivePath, checksumsPath, "moat_Linux_x86_64.tar.gz")

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no checksum")
}

// TestExtractBinary tests the extraction of the moat binary from a
// release archive.
func (suite *UpdateTestSuite) TestExtractBinary() {
	dir := suite.newTestDir()
	archive, err := buildArchive([]byte("new moat binary"))
	require.NoError(suite.T(), err)
	archivePath := filepath.Join(dir, "archive.tar.gz")
	suite.track(archivePath)
	require.NoError(suite.T(), os.WriteFile(archivePath, archive, 0o644))

	binaryPath, err := extractBinary(archivePath, dir)

	require.NoError(suite.T(), err)
	suite.track(binaryPath)
	assert.Equal(suite.T(), []byte("new moat binary"), suite.readFakeBinary(binaryPath))
	info, err := os.Stat(binaryPath)
	require.NoError(suite.T(), err)
	assert.NotZero(suite.T(), info.Mode()&0o111, "the extracted binary should be executable")
}

// TestExtractBinaryWithoutBinary tests that an archive without the moat
// binary is reported.
func (suite *UpdateTestSuite) TestExtractBinaryWithoutBinary() {
	dir := suite.newTestDir()
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)
	require.NoError(suite.T(), tarWriter.WriteHeader(&tar.Header{Name: "LICENSE", Mode: 0o644, Size: 4}))
	_, err := tarWriter.Write([]byte("MIT\n"))
	require.NoError(suite.T(), err)
	require.NoError(suite.T(), tarWriter.Close())
	require.NoError(suite.T(), gzWriter.Close())
	archivePath := filepath.Join(dir, "archive.tar.gz")
	suite.track(archivePath)
	require.NoError(suite.T(), os.WriteFile(archivePath, buf.Bytes(), 0o644))

	_, err = extractBinary(archivePath, dir)

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "does not contain")
}

// TestExtractBinaryDirectoryEntry tests that a directory entry named
// moat is not extracted.
func (suite *UpdateTestSuite) TestExtractBinaryDirectoryEntry() {
	dir := suite.newTestDir()
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)
	require.NoError(suite.T(), tarWriter.WriteHeader(&tar.Header{Name: "moat", Typeflag: tar.TypeDir, Mode: 0o755}))
	require.NoError(suite.T(), tarWriter.Close())
	require.NoError(suite.T(), gzWriter.Close())
	archivePath := filepath.Join(dir, "archive.tar.gz")
	suite.track(archivePath)
	require.NoError(suite.T(), os.WriteFile(archivePath, buf.Bytes(), 0o644))

	_, err := extractBinary(archivePath, dir)

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "not a regular file")
}

// TestDownloadReleaseAssets tests that the archive and the checksums file
// of a release are downloaded.
func (suite *UpdateTestSuite) TestDownloadReleaseAssets() {
	moatSource := suite.newReleaseServer("v0.3.4", []byte("new moat binary"), false)
	suite.setMoatSource(moatSource)
	release, err := fetchLatestRelease()
	require.NoError(suite.T(), err)
	platform := version.PlatformName()

	archivePath, checksumsPath, cleanup, err := downloadReleaseAssets(release, platform, release.TagName)
	require.NoError(suite.T(), err)
	defer cleanup()

	assert.FileExists(suite.T(), archivePath)
	assert.Equal(suite.T(), archiveName(platform), filepath.Base(archivePath))
	assert.FileExists(suite.T(), checksumsPath)
	assert.Equal(suite.T(), checksumsName("v0.3.4"), filepath.Base(checksumsPath))
}

// TestDownloadReleaseAssetsMissingArchive tests that a missing archive
// asset is reported.
func (suite *UpdateTestSuite) TestDownloadReleaseAssetsMissingArchive() {
	release := &Release{
		TagName: "v0.3.4",
		Assets:  []Asset{{Name: "moat_0.3.4_checksums.txt", BrowserDownloadURL: "https://example.com/checksums"}},
	}

	_, _, cleanup, err := downloadReleaseAssets(release, "Linux_x86_64", "v0.3.4")
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no archive")
}

// TestDownloadReleaseAssetsMissingChecksums tests that a missing
// checksums asset is reported.
func (suite *UpdateTestSuite) TestDownloadReleaseAssetsMissingChecksums() {
	release := &Release{
		TagName: "v0.3.4",
		Assets:  []Asset{{Name: "moat_Linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/archive"}},
	}

	_, _, cleanup, err := downloadReleaseAssets(release, "Linux_x86_64", "v0.3.4")
	defer func() {
		if cleanup != nil {
			cleanup()
		}
	}()

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no checksums file")
}

// TestInstallBinaryWithBackup tests that the previous binary is moved to
// the backup path and the new binary replaces the current one.
func (suite *UpdateTestSuite) TestInstallBinaryWithBackup() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	backupPath := filepath.Join(dir, "moat-v0.3.3")
	newBinaryPath := filepath.Join(dir, ".moat.new")
	suite.track(backupPath)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.writeFakeBinary(newBinaryPath, []byte("new binary"))

	require.NoError(suite.T(), installBinary(binaryPath, newBinaryPath, backupPath))

	assert.Equal(suite.T(), []byte("new binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(backupPath))
	assert.Equal(suite.T(), []string{binaryName, "moat-v0.3.3"}, suite.dirEntries(dir))
}

// TestInstallBinaryWithoutBackup tests that the previous binary is
// removed when no backup is requested.
func (suite *UpdateTestSuite) TestInstallBinaryWithoutBackup() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	newBinaryPath := filepath.Join(dir, ".moat.new")
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.writeFakeBinary(newBinaryPath, []byte("new binary"))

	require.NoError(suite.T(), installBinary(binaryPath, newBinaryPath, ""))

	assert.Equal(suite.T(), []byte("new binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestInstallBinaryOverwritesBackup tests that an existing backup file is
// overwritten.
func (suite *UpdateTestSuite) TestInstallBinaryOverwritesBackup() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	backupPath := filepath.Join(dir, "moat-v0.3.3")
	newBinaryPath := filepath.Join(dir, ".moat.new")
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.writeFakeBinary(newBinaryPath, []byte("new binary"))
	suite.writeFakeBinary(backupPath, []byte("stale backup"))

	require.NoError(suite.T(), installBinary(binaryPath, newBinaryPath, backupPath))

	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(backupPath))
}

// TestInstallBinaryFailureRestores tests that the previous binary is
// restored when replacing it fails.
func (suite *UpdateTestSuite) TestInstallBinaryFailureRestores() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	backupPath := filepath.Join(dir, "moat-v0.3.3")
	newBinaryPath := filepath.Join(dir, "missing-new-binary")
	suite.writeFakeBinary(binaryPath, []byte("old binary"))

	err := installBinary(binaryPath, newBinaryPath, backupPath)

	require.Error(suite.T(), err)
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(binaryPath))
	assert.NoFileExists(suite.T(), backupPath)
}

// TestUpdateUpToDate tests that no prompt is asked and the binary is
// left untouched when the running binary is up to date.
func (suite *UpdateTestSuite) TestUpdateUpToDate() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("current binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.newReleaseServer("v0.3.3", []byte("new binary"), false))
	answers := &promptAnswers{updateAnswer: true, backupAnswer: true}
	suite.setPromptAnswers(answers)

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "moat is up to date (v0.3.3)")
	assert.False(suite.T(), answers.updateCalled, "the update prompt should not be asked")
	assert.Equal(suite.T(), []byte("current binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestUpdateWithBackup tests the full update flow with a backup of the
// previous binary.
func (suite *UpdateTestSuite) TestUpdateWithBackup() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), false))
	suite.setPromptAnswers(&promptAnswers{updateAnswer: true, backupAnswer: true})
	suite.track(filepath.Join(dir, "moat-v0.3.3"))

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "Latest release: v0.3.4")
	assert.Contains(suite.T(), output, "Verifying the checksum of the downloaded archive...")
	assert.Contains(suite.T(), output, fmt.Sprintf("Previous binary saved as %s", filepath.Join(dir, "moat-v0.3.3")))
	assert.Contains(suite.T(), output, "moat updated to v0.3.4.")
	assert.Equal(suite.T(), []byte("new binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(filepath.Join(dir, "moat-v0.3.3")))
	info, err := os.Stat(binaryPath)
	require.NoError(suite.T(), err)
	assert.NotZero(suite.T(), info.Mode()&0o111, "the updated binary should be executable")
	assert.Equal(suite.T(), []string{binaryName, "moat-v0.3.3"}, suite.dirEntries(dir))
}

// TestUpdateWithoutBackup tests the full update flow without a backup
// of the previous binary.
func (suite *UpdateTestSuite) TestUpdateWithoutBackup() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), false))
	suite.setPromptAnswers(&promptAnswers{updateAnswer: true, backupAnswer: false})

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "moat updated to v0.3.4.")
	assert.NotContains(suite.T(), output, "Previous binary saved as")
	assert.Equal(suite.T(), []byte("new binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestUpdateDeclined tests that the binary is left untouched when the
// user declines the update.
func (suite *UpdateTestSuite) TestUpdateDeclined() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), false))
	suite.setPromptAnswers(&promptAnswers{updateAnswer: false, backupAnswer: true})

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "Update aborted.")
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestUpdateUnknownCurrentVersion tests that the update still proceeds
// when the current version of the binary is unknown.
func (suite *UpdateTestSuite) TestUpdateUnknownCurrentVersion() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), false))
	suite.setPromptAnswers(&promptAnswers{updateAnswer: true, backupAnswer: true})
	suite.track(filepath.Join(dir, "moat-unknown"))

	output := suite.runUpdate()

	assert.Contains(suite.T(), output, "Could not determine the current version of moat.")
	assert.Contains(suite.T(), output, "Current version: unknown")
	assert.Contains(suite.T(), output, "moat updated to v0.3.4.")
	assert.Equal(suite.T(), []byte("new binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(filepath.Join(dir, "moat-unknown")))
}

// TestUpdateChecksumMismatch tests that the binary is left untouched
// when the checksum of the downloaded archive does not match.
func (suite *UpdateTestSuite) TestUpdateChecksumMismatch() {
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.newReleaseServer("v0.3.4", []byte("new binary"), true))
	suite.setPromptAnswers(&promptAnswers{updateAnswer: true, backupAnswer: true})

	capture := utils.OutputCapture{}
	capture.StartCapture()
	err := Update()
	_, captureErr := capture.StopCapture()
	require.NoError(suite.T(), captureErr)

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "checksum verification failed")
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(binaryPath))
	assert.Equal(suite.T(), []string{binaryName}, suite.dirEntries(dir))
}

// TestUpdateFetchFailure tests that an error is returned when the latest
// release cannot be fetched.
func (suite *UpdateTestSuite) TestUpdateFetchFailure() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	suite.T().Cleanup(server.Close)
	dir := suite.newTestDir()
	binaryPath := filepath.Join(dir, binaryName)
	suite.writeFakeBinary(binaryPath, []byte("old binary"))
	suite.setExecutablePath(binaryPath)
	suite.setVersion("v0.3.3")
	suite.setMoatSource(suite.testSource(server))

	capture := utils.OutputCapture{}
	capture.StartCapture()
	err := Update()
	_, captureErr := capture.StopCapture()
	require.NoError(suite.T(), captureErr)

	require.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "fetching the latest release failed")
	assert.Equal(suite.T(), []byte("old binary"), suite.readFakeBinary(binaryPath))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUpdateTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateTestSuite))
}
