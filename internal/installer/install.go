package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
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

// installDeps holds injectable dependencies for installation helpers.
// Production code uses defaultInstallDeps(); tests pass explicit values.
type installDeps struct {
	client         *http.Client
	allowLocalhost bool
	allowCustom    bool
}

func defaultInstallDeps() installDeps {
	return installDeps{
		client: newScriptHTTPClient(),
	}
}

func newScriptHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("too many redirects")
			}
			return validateURL(req.URL.String(), false)
		},
	}
}

// ProgressFn is a callback type for reporting installation progress.
// The agentID identifies which agent is being installed, and percentage
// is a value between 0 and 100 indicating estimated completion.
//
// Conventional percentages:
//   - 0: installation of this agent is starting (may download large binaries)
//   - 50: install script verified; running the installer command
//   - 100: installation finished successfully
//
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
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating log directory: %w", err)
	}
	return dir, nil
}

// logPath generates a log file path for the given agent ID. The filename
// includes a timestamp to prevent collisions across runs. Colons in the
// timestamp are replaced with hyphens for filesystem compatibility.
// Returns the full path and any error from ensuring the directory exists.
func logPath(agentID string) (string, error) {
	if err := validateAgentID(agentID); err != nil {
		return "", err
	}
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
// For curl_bash methods, the install script is downloaded once to a temp
// file, verified against Checksum.SHA256, and then executed from that file
// so the verified bytes are the executed bytes. Installation is refused if
// no checksum is provided for a curl_bash method.
//
// npm_install and custom methods run as argv (no shell), after validation.
//
// On success, progress(agent.ID, 100) is called. progress(agent.ID, 0) is
// called when work begins so UIs can show activity during long downloads.
// On failure, the error is wrapped with additional context and returned.
func InstallAgent(agent registry.Agent, progress ProgressFn) error {
	return installAgent(agent, progress, defaultInstallDeps())
}

func installAgent(agent registry.Agent, progress ProgressFn, deps installDeps) error {
	if agent.Install.Method == registry.MethodCustom && !deps.allowCustom {
		return fmt.Errorf("installing %s: custom install method is not allowed from registry data", agent.ID)
	}
	if err := validateAgentInstall(agent); err != nil {
		return err
	}

	reportProgress(progress, agent.ID, 0)

	name, args, cleanup, err := buildInstallArgv(agent, deps)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	reportProgress(progress, agent.ID, 50)

	// Include agent home bin dir so installers that write to ~/.opencode/bin
	// (and only update interactive ~/.bashrc) still run and detect correctly.
	extraBin := agentHomeBinDir(agent.ID)
	if err := runAndLog(agent.ID, name, args, nil, extraBin); err != nil {
		return err
	}
	if agent.DetectCmd != "" && !AgentIsInstalled(agent) {
		return fmt.Errorf("installing %s: command succeeded but %s binary not found in PATH", agent.ID, agent.DetectCmd)
	}
	reportProgress(progress, agent.ID, 100)
	return nil
}

func reportProgress(progress ProgressFn, agentID string, pct int) {
	if progress != nil {
		progress(agentID, pct)
	}
}

func validateAgentInstall(agent registry.Agent) error {
	if err := validateAgentID(agent.ID); err != nil {
		return fmt.Errorf("installing %s: %w", agent.ID, err)
	}
	if agent.Install.Command == "" {
		return fmt.Errorf("installing %s: install command is empty", agent.ID)
	}
	if strings.ContainsRune(agent.Install.Command, '\x00') {
		return fmt.Errorf("installing %s: install command contains null byte", agent.ID)
	}
	if err := validateCommand(agent.Install.Method, agent.Install.Command); err != nil {
		return fmt.Errorf("installing %s: %w", agent.ID, err)
	}
	return nil
}

// buildInstallArgv resolves the executable and args for an install method.
// curl_bash downloads+verifies a script and returns [shell, scriptPath].
// Other methods parse the command into argv without invoking a shell.
func buildInstallArgv(agent registry.Agent, deps installDeps) (name string, args []string, cleanup func(), err error) {
	switch agent.Install.Method {
	case registry.MethodCurlBash:
		return prepareCurlBashArgv(agent, deps)
	case registry.MethodNpmInstall, registry.MethodCustom:
		fields := strings.Fields(agent.Install.Command)
		if len(fields) == 0 {
			return "", nil, nil, fmt.Errorf("installing %s: empty command after parsing", agent.ID)
		}
		return fields[0], fields[1:], nil, nil
	default:
		return "", nil, nil, fmt.Errorf("installing %s: unknown install method %q", agent.ID, agent.Install.Method)
	}
}

// prepareCurlBashArgv downloads the install script once, verifies its
// checksum, and returns argv that executes the verified file.
func prepareCurlBashArgv(agent registry.Agent, deps installDeps) (name string, args []string, cleanup func(), err error) {
	scriptPath, err := downloadAndVerifyScript(agent.ID, agent.Install.URL, agent.Checksum, deps)
	if err != nil {
		return "", nil, nil, err
	}
	cleanup = func() { removeTempScript(scriptPath) }

	shell := pipelineShell(agent.Install.Command)
	if agent.Install.NonInteractive {
		// Feed yes into the verified script via shell-less process group:
		// run `shell scriptPath` with stdin from `yes` using a pipe in runAndLog
		// is more complex; keep a tiny wrapper argv via `sh -c` only for the
		// verified local path (no remote content interpolation).
		return "sh", []string{"-c", "yes | " + shell + " \"$1\"", "squad-install", scriptPath}, cleanup, nil
	}
	return shell, []string{scriptPath}, cleanup, nil
}

// pipelineShell extracts the shell interpreter from a curl|shell pipeline.
// Defaults to bash when the trailing command is empty or unrecognized.
func pipelineShell(command string) string {
	lastPipe := strings.LastIndex(command, "|")
	if lastPipe < 0 {
		return "bash"
	}
	after := strings.TrimSpace(command[lastPipe+1:])
	fields := strings.Fields(after)
	if len(fields) == 0 {
		return "bash"
	}
	base := filepath.Base(fields[0])
	switch base {
	case "bash", "sh", "dash", "zsh":
		return base
	default:
		return "bash"
	}
}

// runAndLog executes name+args and writes combined output to a log file.
// Returns a wrapped error on failure. progress is reserved for callers that
// still pass it (uninstall); install flow reports progress outside runAndLog.
// extraBinDirs are prepended to PATH for the child process only.
func runAndLog(agentID, name string, args []string, progress ProgressFn, extraBinDirs ...string) error {
	path, err := logPath(agentID)
	if err != nil {
		return fmt.Errorf("preparing log path for %s: %w", agentID, err)
	}

	cmd := exec.Command(name, args...)
	// Installers often drop binaries into ~/.local/bin and similar paths that
	// are missing from non-interactive PATH. Keep the child env consistent
	// with post-install detection without mutating process-global state.
	cmd.Env = append(os.Environ(), "PATH="+pathWithUserBins(extraBinDirs...))
	output, err := cmd.CombinedOutput()

	if writeErr := os.WriteFile(path, output, 0o644); writeErr != nil {
		// Prefer the install failure if both happened.
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				return fmt.Errorf("installing %s: command exited with code %d (also failed writing log: %v)", agentID, exitErr.ExitCode(), writeErr)
			}
			return fmt.Errorf("installing %s: %w (also failed writing log: %v)", agentID, err, writeErr)
		}
		return fmt.Errorf("installing %s: writing log: %w", agentID, writeErr)
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return fmt.Errorf("installing %s: command exited with code %d (see log: %s)", agentID, exitErr.ExitCode(), path)
		}
		return fmt.Errorf("installing %s: %w (see log: %s)", agentID, err, path)
	}

	reportProgress(progress, agentID, 100)
	return nil
}

// downloadAndVerifyScript downloads the script from url into a temp file,
// verifies its SHA-256 against the expected checksum, and returns the path.
// The caller owns the file and must remove it.
func downloadAndVerifyScript(agentID, scriptURL string, cksum *registry.Checksum, deps installDeps) (string, error) {
	if cksum == nil || cksum.SHA256 == "" {
		return "", fmt.Errorf("installing %s: checksum required for curl_bash install method", agentID)
	}
	if err := validateURL(scriptURL, deps.allowLocalhost); err != nil {
		return "", fmt.Errorf("installing %s: %w", agentID, err)
	}

	body, err := fetchScriptBody(agentID, scriptURL, deps.client)
	if err != nil {
		return "", err
	}
	return writeVerifiedScript(agentID, body, cksum.SHA256)
}

func fetchScriptBody(agentID, scriptURL string, client *http.Client) ([]byte, error) {
	if client == nil {
		client = defaultInstallDeps().client
	}
	req, err := http.NewRequest(http.MethodGet, scriptURL, nil)
	if err != nil {
		return nil, fmt.Errorf("installing %s: building download request: %w", agentID, err)
	}
	req.Header.Set("User-Agent", "squad-ai/"+agentID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("installing %s: downloading script for checksum verification: %w", agentID, err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close after full read

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("installing %s: script download returned status %d", agentID, resp.StatusCode)
	}
	const maxScriptBytes = 5 << 20 // 5 MiB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxScriptBytes+1))
	if err != nil {
		return nil, fmt.Errorf("installing %s: reading script content: %w", agentID, err)
	}
	if len(body) > maxScriptBytes {
		return nil, fmt.Errorf("installing %s: script exceeds %d byte limit", agentID, maxScriptBytes)
	}
	return body, nil
}

func writeVerifiedScript(agentID string, body []byte, expectedSHA string) (string, error) {
	got := hex.EncodeToString(sha256Sum(body))
	if got != expectedSHA {
		return "", fmt.Errorf("installing %s: checksum mismatch (expected %s, got %s)", agentID, expectedSHA, got)
	}

	tmp, err := os.CreateTemp("", "squad-install-*.sh")
	if err != nil {
		return "", fmt.Errorf("installing %s: creating temp script file: %w", agentID, err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(body); err != nil {
		closeIgnore(tmp)
		removeTempScript(tmpPath)
		return "", fmt.Errorf("installing %s: writing temp script file: %w", agentID, err)
	}
	if err := tmp.Close(); err != nil {
		removeTempScript(tmpPath)
		return "", fmt.Errorf("installing %s: closing temp script file: %w", agentID, err)
	}
	if err := os.Chmod(tmpPath, 0o700); err != nil {
		removeTempScript(tmpPath)
		return "", fmt.Errorf("installing %s: setting temp script permissions: %w", agentID, err)
	}
	return tmpPath, nil
}

func sha256Sum(body []byte) []byte {
	sum := sha256.Sum256(body)
	return sum[:]
}

func removeTempScript(path string) {
	// Best-effort cleanup; the primary error must not be masked.
	_ = os.Remove(path)
}

func closeIgnore(c interface{ Close() error }) {
	if c == nil {
		return
	}
	// Best-effort close on error paths; the primary error is returned to the caller.
	_ = c.Close()
}

// validateAgentID restricts agent IDs used in filesystem paths.
func validateAgentID(id string) error {
	if id == "" {
		return errors.New("agent id is empty")
	}
	if len(id) > 64 {
		return errors.New("agent id is too long")
	}
	for i, r := range id {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if i == 0 && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return errors.New("agent id must start with a lowercase letter or digit")
		}
		if !ok {
			return errors.New("agent id must match [a-z0-9-]+")
		}
	}
	return nil
}

// validateURL checks that u is an HTTPS URL targeting a remote host.
func validateURL(rawURL string, allowLocalhost bool) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return errors.New("URL must use HTTPS scheme")
	}
	if u.Host == "" {
		return errors.New("URL must have a host")
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		if allowLocalhost {
			return nil
		}
		return errors.New("URL must not target localhost")
	}
	return nil
}

// validateCommand checks that the install command matches the expected
// pattern for its install method before execution.
func validateCommand(method registry.InstallMethod, command string) error {
	switch method {
	case registry.MethodCurlBash:
		if !strings.HasPrefix(command, "curl ") {
			return errors.New("curl_bash command must start with 'curl '")
		}
		if !strings.Contains(command, "|") {
			return errors.New("curl_bash command must contain a pipe")
		}
	case registry.MethodNpmInstall:
		fields := strings.Fields(command)
		if len(fields) < 2 || fields[0] != "npm" {
			return errors.New("npm_install command must be an npm argv starting with npm")
		}
	case registry.MethodCustom:
		// Custom is intended for local tests/helpers: absolute binary path + args.
		fields := strings.Fields(command)
		if len(fields) == 0 {
			return errors.New("custom command must not be empty")
		}
		if !strings.HasPrefix(fields[0], "/") {
			return errors.New("custom command must be an absolute path")
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
	}
	return results
}
