package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alebak/squad-ai/internal/registry"
)

// ProgressFn is a callback type for reporting installation progress.
// The agentID identifies which agent is being installed, and percentage
// is a value between 0 and 100 indicating estimated completion.
// Callers should check for nil before calling.
type ProgressFn func(agentID string, percentage int)

// logDir returns the path to the logs directory under the user's config
// directory using os.UserConfigDir(). It creates the directory with 0755
// permissions if it does not exist.
func logDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}

	dir := filepath.Join(cfgDir, "squad-ai", "logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("creating log directory: %w", err)
	}
	return dir, nil
}

// logPath generates a log file path for the given agent ID. The filename
// includes a timestamp to prevent collisions across runs. Colons in the
// timestamp are replaced with hyphens for filesystem compatibility.
// Returns the full path and any error from ensuring the directory exists.
func logPath(agentID string) (string, error) {
	dir, err := logDir()
	if err != nil {
		return "", err
	}

	ts := time.Now().Format(time.RFC3339)
	ts = strings.ReplaceAll(ts, ":", "-")
	return filepath.Join(dir, fmt.Sprintf("%s-%s.log", agentID, ts)), nil
}

// InstallAgent executes an agent's install command, captures its combined
// stdout and stderr to a log file, and reports progress via the callback.
//
// The command is run via "sh -c <command>" to support pipes and shell
// syntax used by typical agent install scripts (curl | bash, npm install -g).
//
// For curl_bash install methods, the agent's Checksum.SHA256 is verified
// against the content at agent.Install.URL before execution. Installation
// is refused if no checksum is provided for a curl_bash method.
//
// On success, progress(agent.ID, 100) is called. On failure, the error
// is wrapped with additional context and returned.
func InstallAgent(agent registry.Agent, progress ProgressFn) error {
	if agent.Install.Command == "" {
		return fmt.Errorf("installing %s: install command is empty", agent.ID)
	}
	if strings.ContainsRune(agent.Install.Command, '\x00') {
		return fmt.Errorf("installing %s: install command contains null byte", agent.ID)
	}
	if err := validateCommand(agent.Install.Method, agent.Install.Command); err != nil {
		return fmt.Errorf("installing %s: %w", agent.ID, err)
	}
	if agent.Install.Method == registry.MethodCurlBash {
		if err := verifyChecksum(agent.ID, agent.Install.URL, agent.Checksum); err != nil {
			return err
		}
	}
	if err := runAndLog(agent.ID, agent.Install.Command, progress); err != nil {
		return err
	}
	return nil
}

// runAndLog executes cmd via sh -c, writes output to a log file, and
// reports progress on success. Returns a wrapped error on failure.
func runAndLog(agentID, command string, progress ProgressFn) error {
	path, err := logPath(agentID)
	if err != nil {
		return fmt.Errorf("preparing log path for %s: %w", agentID, err)
	}

	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()

	if writeErr := os.WriteFile(path, output, 0644); writeErr != nil {
		return fmt.Errorf("installing %s: writing log: %w", agentID, writeErr)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("installing %s: command exited with code %d (see log: %s)", agentID, exitErr.ExitCode(), path)
		}
		return fmt.Errorf("installing %s: %w (see log: %s)", agentID, err, path)
	}

	if progress != nil {
		progress(agentID, 100)
	}
	return nil
}

// verifyChecksum downloads the script from url and compares its SHA-256
// hash against the expected checksum. The URL is validated before the
// request: it must use HTTPS and target a public host.
func verifyChecksum(agentID, scriptURL string, cksum *registry.Checksum) error {
	if cksum == nil || cksum.SHA256 == "" {
		return fmt.Errorf("installing %s: checksum required for curl_bash install method", agentID)
	}
	if err := validateURL(scriptURL); err != nil {
		return fmt.Errorf("installing %s: %w", agentID, err)
	}

	resp, err := http.Get(scriptURL)
	if err != nil {
		return fmt.Errorf("installing %s: downloading script for checksum verification: %w", agentID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("installing %s: script download returned status %d", agentID, resp.StatusCode)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, resp.Body); err != nil {
		return fmt.Errorf("installing %s: hashing script content: %w", agentID, err)
	}

	got := hex.EncodeToString(hasher.Sum(nil))
	if got != cksum.SHA256 {
		return fmt.Errorf("installing %s: checksum mismatch (expected %s, got %s)", agentID, cksum.SHA256, got)
	}

	return nil
}

// validateURL checks that u is an HTTPS URL targeting a remote host.
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS scheme")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return fmt.Errorf("URL must not target localhost")
	}
	return nil
}

// validateCommand checks that the install command matches the expected
// pattern for its install method. This validates data from the remote
// registry before it reaches the shell. For MethodCustom, any non-empty
// command is accepted.
func validateCommand(method registry.InstallMethod, command string) error {
	switch method {
	case registry.MethodCurlBash:
		if !strings.HasPrefix(command, "curl ") {
			return fmt.Errorf("curl_bash command must start with 'curl '")
		}
		if !strings.Contains(command, "|") {
			return fmt.Errorf("curl_bash command must contain a pipe")
		}
	case registry.MethodNpmInstall:
		if !strings.HasPrefix(command, "npm ") {
			return fmt.Errorf("npm_install command must start with 'npm '")
		}
	case registry.MethodCustom:
		if len(command) == 0 {
			return fmt.Errorf("custom command must not be empty")
		}
	default:
		return fmt.Errorf("unknown install method %q", method)
	}
	return nil
}

// InstallAll installs all provided agents sequentially and returns a slice
// of errors with the same length as the input. A nil entry means the agent
// was installed successfully. InstallAll does NOT stop on the first failure —
// it attempts every agent and collects all results.
func InstallAll(agents []registry.Agent, progress ProgressFn) []error {
	if len(agents) == 0 {
		return nil
	}

	results := make([]error, len(agents))
	for i, agent := range agents {
		err := InstallAgent(agent, progress)
		if err != nil {
			results[i] = fmt.Errorf("[%d] %w", i, err)
		}
		// results[i] stays nil on success
	}
	return results
}


