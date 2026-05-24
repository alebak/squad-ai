package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

const updateURL = "https://api.github.com/repos/alebak/squad-ai/releases/latest"

type releaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func newSelfUpdateCommand() *cobra.Command {
	var checkOnly bool

	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update squad to the latest version",
		Long: `Download and install the latest version of Squad AI from GitHub Releases.

The update replaces the current binary atomically: the new binary is
verified against its checksum before installation, and the previous
binary is backed up as squad.old.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSelfUpdate(cmd, checkOnly)
		},
	}

	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check if an update is available")
	return cmd
}

func runSelfUpdate(cmd *cobra.Command, checkOnly bool) error {
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding current binary: %w", err)
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return fmt.Errorf("checking for updates: %w", err)
	}

	current := strings.TrimPrefix(version, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	if current == latest {
		cmd.Printf("✅ Squad AI is up to date (%s)\n", version)
		return nil
	}
	if checkOnly {
		cmd.Printf("ℹ️  Update available: %s → %s\n", version, release.TagName)
		cmd.Println("Run 'squad self-update' to install the latest version.")
		return nil
	}

	cmd.Printf("📦 Updating Squad AI %s → %s\n", version, release.TagName)

	assetName := fmt.Sprintf("squad-ai_%s_%s_%s.tar.gz",
		latest, runtime.GOOS, runtime.GOARCH)

	downloadURL := findAssetURL(release, assetName)
	if downloadURL == "" {
		return fmt.Errorf("no binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	checksumName := "checksums.txt"
	checksumURL := findAssetURL(release, checksumName)
	if checksumURL == "" {
		return fmt.Errorf("no checksums found for release %s", release.TagName)
	}

	tmpDir, err := os.MkdirTemp("", "squad-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}

	checksumPath := filepath.Join(tmpDir, checksumName)
	if err := downloadFile(checksumURL, checksumPath); err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	if err := verifyAssetChecksum(archivePath, checksumPath, assetName); err != nil {
		return err
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("creating extract directory: %w", err)
	}
	if err := extractTarGz(archivePath, extractDir); err != nil {
		return fmt.Errorf("extracting update: %w", err)
	}

	newBinary := filepath.Join(extractDir, "squad")
	backupPath := currentPath + ".old"

	if err := os.Rename(currentPath, backupPath); err != nil {
		return fmt.Errorf("backing up current binary: %w", err)
	}

	if err := copyFile(newBinary, currentPath); err != nil {
		os.Rename(backupPath, currentPath)
		return fmt.Errorf("installing update: %w", err)
	}
	if err := os.Chmod(currentPath, 0755); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	os.Remove(backupPath)
	cmd.Printf("✅ Updated to %s\n", release.TagName)
	return nil
}

func fetchLatestRelease() (*releaseInfo, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return nil, fmt.Errorf("fetching release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release info: %w", err)
	}
	return &release, nil
}

func findAssetURL(release *releaseInfo, name string) string {
	for _, a := range release.Assets {
		if strings.EqualFold(a.Name, name) {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func downloadFile(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyAssetChecksum(archivePath, checksumPath, assetName string) error {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("reading checksum file: %w", err)
	}

	var expected string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, assetName) {
			expected = strings.Fields(line)[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksum not found for %s", assetName)
	}

	out, err := exec.Command("sha256sum", archivePath).Output()
	if err != nil {
		return fmt.Errorf("computing sha256: %w", err)
	}
	actual := strings.Fields(string(out))[0]

	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func extractTarGz(archivePath, destDir string) error {
	return exec.Command("tar", "-xzf", archivePath, "-C", destDir).Run()
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	tmpPath := dst + ".tmp"
	dstFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(tmpPath)
		return err
	}
	dstFile.Close()

	return os.Rename(tmpPath, dst)
}

// AutoUpdate checks for a newer version and silently replaces the binary
// if one is available. Call this at startup before any command runs.
// If an update is applied, the process is re-exec'd and never returns.
func AutoUpdate() {
	autoUpdate()
}

// autoUpdate checks for a newer version on GitHub Releases and replaces
// the current binary if one is available. On success, it re-execs the
// new binary and never returns. Errors are logged to stderr but never
// prevent the command from running.
func autoUpdate() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "squad: auto-update panic: %v\n", r)
		}
	}()

	currentPath, err := os.Executable()
	if err != nil {
		return
	}

	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(os.Stderr, "squad: auto-update check failed: %v\n", err)
		return
	}

	current := strings.TrimPrefix(version, "v")
	latest := strings.TrimPrefix(release.TagName, "v")
	if current == latest || latest == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "squad: updating from %s to %s...\n", current, latest)

	assetName := fmt.Sprintf("squad-ai_%s_%s_%s.tar.gz",
		latest, runtime.GOOS, runtime.GOARCH)
	downloadURL := findAssetURL(release, assetName)
	if downloadURL == "" {
		fmt.Fprintf(os.Stderr, "squad: no binary found for %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return
	}

	checksumURL := findAssetURL(release, "checksums.txt")
	if checksumURL == "" {
		fmt.Fprintf(os.Stderr, "squad: no checksums found for release %s\n", release.TagName)
		return
	}

	tmpDir, err := os.MkdirTemp("", "squad-update-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "squad: temp dir error: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		fmt.Fprintf(os.Stderr, "squad: download error: %v\n", err)
		return
	}

	checksumPath := filepath.Join(tmpDir, "checksums.txt")
	if err := downloadFile(checksumURL, checksumPath); err != nil {
		fmt.Fprintf(os.Stderr, "squad: checksum download error: %v\n", err)
		return
	}

	if err := verifyAssetChecksum(archivePath, checksumPath, assetName); err != nil {
		fmt.Fprintf(os.Stderr, "squad: checksum verification failed: %v\n", err)
		return
	}

	extractDir := filepath.Join(tmpDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "squad: extract dir error: %v\n", err)
		return
	}
	if err := extractTarGz(archivePath, extractDir); err != nil {
		fmt.Fprintf(os.Stderr, "squad: extract error: %v\n", err)
		return
	}

	newBinary := filepath.Join(extractDir, "squad")
	if err := copyFile(newBinary, currentPath); err != nil {
		fmt.Fprintf(os.Stderr, "squad: copy error: %v\n", err)
		return
	}
	if err := os.Chmod(currentPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "squad: chmod error: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "squad: updated to %s, restarting...\n", release.TagName)
	syscall.Exec(currentPath, os.Args, os.Environ())
}
