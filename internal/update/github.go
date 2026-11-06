package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AaltoRSE/moat/internal/utils"
	"github.com/AaltoRSE/moat/internal/version"
)

// Release describes a release of the upstream moat repository as reported
// by the GitHub API. Only the fields used by the update are included.
type Release struct {
	// TagName is the git tag of the release, for example v0.3.3.
	TagName string `json:"tag_name"`
	// Assets are the files attached to the release.
	Assets []Asset `json:"assets"`
}

// Asset is a single file attached to a release.
type Asset struct {
	// Name is the file name of the asset, for example
	// moat_Linux_x86_64.tar.gz.
	Name string `json:"name"`
	// BrowserDownloadURL is the URL from which the asset can be
	// downloaded.
	BrowserDownloadURL string `json:"browser_download_url"`
}

const (
	// apiTimeout limits a single request to the GitHub API.
	apiTimeout = 30 * time.Second

	// downloadTimeout limits a single asset download.
	downloadTimeout = 5 * time.Minute

	// userAgent identifies moat to the GitHub API and the download
	// server.
	userAgent = "moat-update"
)

// getAPIUrl returns the URL of the GitHub API endpoint that reports the
// latest release of the repository at the given source, for example
// https://api.github.com/repos/AaltoRSE/moat/releases/latest for the
// source github.com/AaltoRSE/moat. A "http://" or "https://" prefix and
// a ".git" suffix of the source are ignored, and trailing slashes are
// trimmed. A source without a host defaults to github.com, the host
// github.com is mapped to api.github.com, and other hosts are used as
// they are. The scheme is https unless the source explicitly uses
// "http://". It returns an error when the source is not a GitHub
// repository (an owner and a repository name, optionally preceded by a
// host).
func getAPIUrl(source string) (string, error) {
	scheme := "https"
	rest := source
	if strings.HasPrefix(rest, "https://") {
		rest = strings.TrimPrefix(rest, "https://")
	} else if strings.HasPrefix(rest, "http://") {
		rest = strings.TrimPrefix(rest, "http://")
		scheme = "http"
	}
	rest = strings.TrimSuffix(rest, ".git")
	rest = strings.Trim(rest, "/")

	segments := strings.Split(rest, "/")
	host := "github.com"
	switch len(segments) {
	case 2: // owner/repo
	case 3: // host/owner/repo
		host, segments = segments[0], segments[1:]
	default:
		return "", fmt.Errorf("the source %q is not a GitHub repository", source)
	}
	if host == "github.com" {
		host = "api.github.com"
	}
	return fmt.Sprintf("%s://%s/repos/%s/releases/latest", scheme, host, strings.Join(segments, "/")), nil
}

// fetchLatestRelease returns the latest release of the upstream moat
// repository as reported by the GitHub API endpoint of the source
// repository [version.MoatSource] (see [getAPIUrl]). It returns an error
// when the endpoint cannot be derived from the source, the request
// fails, or the response does not contain release information.
func fetchLatestRelease() (*Release, error) {
	apiURL, err := getAPIUrl(version.MoatSource)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: apiTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("request for the latest release failed: %s (%s)", resp.Status, strings.TrimSpace(string(body)))
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing the release information failed: %w", err)
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("the release information contains no tag name")
	}
	return &release, nil
}

// asset returns the release asset whose file name is name, or nil when
// the release has no such asset.
func (r *Release) asset(name string) *Asset {
	for i := range r.Assets {
		if r.Assets[i].Name == name {
			return &r.Assets[i]
		}
	}
	return nil
}

// archiveName returns the file name of the release archive asset for the
// given platform name (see [version.PlatformName]), for example
// moat_Linux_x86_64.tar.gz.
func archiveName(platform string) string {
	return fmt.Sprintf("%s_%s.tar.gz", binaryName, platform)
}

// checksumsName returns the file name of the checksums asset for the
// given release tag, for example moat_0.3.3_checksums.txt for the tag
// v0.3.3.
func checksumsName(tag string) string {
	return fmt.Sprintf("%s_%s_checksums.txt", binaryName, strings.TrimPrefix(tag, "v"))
}

// downloadReleaseAssets downloads the release archive and the checksums
// file of the given release for the platform name (see
// [version.PlatformName]) into a new temporary directory. It returns the
// paths of the downloaded archive and checksums file, and a cleanup
// function that removes the downloaded files one by one and then removes
// the temporary directory. It returns an error when the release does not
// contain the required assets or a download fails; the downloaded files
// and the temporary directory are removed in that case as well.
func downloadReleaseAssets(release *Release, platform, tag string) (archivePath, checksumsPath string, cleanup func(), err error) {
	archive := release.asset(archiveName(platform))
	if archive == nil {
		return "", "", nil, fmt.Errorf("release %s contains no archive for %s", release.TagName, platform)
	}
	checksums := release.asset(checksumsName(tag))
	if checksums == nil {
		return "", "", nil, fmt.Errorf("release %s contains no checksums file", release.TagName)
	}

	dir, err := os.MkdirTemp("", "moat-update-")
	if err != nil {
		return "", "", nil, err
	}
	// downloaded holds the paths of the files downloaded (or partially
	// downloaded) so far. Each path is added before its download starts,
	// so that the cleanup function also removes the remains of failed
	// downloads.
	var downloaded []string
	cleanup = func() {
		_ = utils.CleanupFiles(downloaded)
		_ = utils.CleanupDir(dir)
	}

	archivePath = filepath.Join(dir, archive.Name)
	downloaded = append(downloaded, archivePath)
	if err := downloadFile(archive.BrowserDownloadURL, archivePath); err != nil {
		cleanup()
		return "", "", nil, fmt.Errorf("downloading %s failed: %w", archive.Name, err)
	}
	checksumsPath = filepath.Join(dir, checksums.Name)
	downloaded = append(downloaded, checksumsPath)
	if err := downloadFile(checksums.BrowserDownloadURL, checksumsPath); err != nil {
		cleanup()
		return "", "", nil, fmt.Errorf("downloading %s failed: %w", checksums.Name, err)
	}
	return archivePath, checksumsPath, cleanup, nil
}

// downloadFile downloads the file at url and writes its contents to
// destPath. It returns an error when the request fails or does not return
// HTTP 200.
func downloadFile(url, destPath string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: downloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
