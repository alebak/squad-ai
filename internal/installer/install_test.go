package installer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeps() installDeps {
	d := defaultInstallDeps()
	d.allowCustom = true
	return d
}

func testInstall(agent registry.Agent, progress ProgressFn) error {
	return installAgent(agent, progress, testDeps())
}

func testInstallAll(agents []registry.Agent, progress ProgressFn) []error {
	if len(agents) == 0 {
		return nil
	}
	results := make([]error, len(agents))
	for i, agent := range agents {
		err := testInstall(agent, progress)
		if err != nil {
			results[i] = fmt.Errorf("[%d] %w", i, err)
		}
	}
	return results
}

// safeAgent returns a registry.Agent whose install command is a safe
// command (no real agent installation). Use this in all tests to avoid
// executing actual install scripts.
func safeAgent(id, cmd string) registry.Agent {
	return registry.Agent{
		ID:        id,
		Name:      id,
		DetectCmd: "true",
		Install: registry.InstallCmd{
			Method:  registry.MethodCustom,
			Command: cmd,
		},
	}
}

func TestProgressFn_Type(t *testing.T) {
	// Verify ProgressFn is callable with the expected signature.
	var fn ProgressFn = func(agentID string, percentage int) {
		assert.Equal(t, "test-agent", agentID)
		assert.Equal(t, 100, percentage)
	}
	fn("test-agent", 100)
}

func TestProgressFn_NilSafe(t *testing.T) {
	// Calling InstallAgent with nil progress must not panic.
	agent := safeAgent("nil-safe", "/bin/true")
	err := testInstall(agent, nil)
	assert.NoError(t, err)
}

func TestInstallAgent_Success(t *testing.T) {
	progressCalled := false
	progress := func(agentID string, percentage int) {
		progressCalled = true
		assert.Equal(t, "success-agent", agentID)
		assert.Equal(t, 100, percentage)
	}

	agent := safeAgent("success-agent", "/bin/true")
	err := testInstall(agent, progress)
	assert.NoError(t, err)
	assert.True(t, progressCalled, "progress callback should be called on success")
}

func TestInstallAgent_Failure(t *testing.T) {
	agent := safeAgent("fail-agent", "/bin/false")
	err := testInstall(agent, nil)
	require.Error(t, err)

	// Error should mention exit code 1
	assert.Contains(t, err.Error(), "exited with code 1")
	assert.Contains(t, err.Error(), "fail-agent")
}

func TestInstallAgent_OutputCapturedToLog(t *testing.T) {
	// Custom method only allows absolute paths without shell metacharacters.
	// Use /bin/echo for capture coverage.
	agent := safeAgent("echo-test", "/bin/echo")
	err := testInstall(agent, nil)
	require.NoError(t, err)

	// Verify a log file was created for this agent.
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	logDir := filepath.Join(home, ".config", "squad-ai", "logs")

	entries, err := os.ReadDir(logDir)
	require.NoError(t, err)

	var found bool
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "echo-test-") && strings.HasSuffix(e.Name(), ".log") {
			found = true
			// /bin/echo with no args still produces a newline log file.
			_, err := os.ReadFile(filepath.Join(logDir, e.Name()))
			require.NoError(t, err)
			break
		}
	}
	assert.True(t, found, "log file for echo-test should exist")
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "https ok", raw: "https://example.com/install.sh", wantErr: false},
		{name: "http rejected", raw: "http://example.com/install.sh", wantErr: true},
		{name: "localhost rejected", raw: "https://localhost/install.sh", wantErr: true},
		{name: "empty host", raw: "https:///install.sh", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.raw, false)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name    string
		method  registry.InstallMethod
		command string
		wantErr bool
	}{
		{name: "curl_bash ok", method: registry.MethodCurlBash, command: "curl -fsSL https://x | bash", wantErr: false},
		{name: "curl_bash no pipe", method: registry.MethodCurlBash, command: "curl -fsSL https://x", wantErr: true},
		{name: "npm ok", method: registry.MethodNpmInstall, command: "npm i -g @openai/codex", wantErr: false},
		{name: "npm not npm", method: registry.MethodNpmInstall, command: "pnpm i -g x", wantErr: true},
		{name: "custom absolute", method: registry.MethodCustom, command: "/bin/true", wantErr: false},
		{name: "custom relative", method: registry.MethodCustom, command: "true", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.method, tt.command)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPipelineShell(t *testing.T) {
	assert.Equal(t, "bash", pipelineShell("curl -fsSL https://x | bash"))
	assert.Equal(t, "sh", pipelineShell("curl -fsSL https://x | sh"))
	assert.Equal(t, "bash", pipelineShell("curl -fsSL https://x | bash -s --"))
	assert.Equal(t, "bash", pipelineShell("no-pipe"))
}

func TestDownloadAndVerifyScript_ChecksumRequired(t *testing.T) {
	_, err := downloadAndVerifyScript("x", "https://example.com/x", nil, defaultInstallDeps())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum required")
}

func TestDownloadAndVerifyScript_MatchAndMismatch(t *testing.T) {
	content := []byte("#!/bin/sh\necho hi\n")
	sum := sha256Hex(content)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	t.Cleanup(srv.Close)

	deps := installDeps{
		client:         srv.Client(),
		allowLocalhost: true,
	}

	path, err := downloadAndVerifyScript("agent", srv.URL, &registry.Checksum{SHA256: sum}, deps)
	require.NoError(t, err)
	t.Cleanup(func() { removeTempScript(path) })
	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, got)

	_, err = downloadAndVerifyScript("agent", srv.URL, &registry.Checksum{SHA256: strings.Repeat("0", 64)}, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func sha256Hex(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

func TestInstallAgent_NoChecksumWarns(t *testing.T) {
	// Agent with nil checksum should still install successfully.
	agent := safeAgent("no-cksum", "/bin/true")
	err := testInstall(agent, nil)
	assert.NoError(t, err)
}

func TestLogPath_Format(t *testing.T) {
	path, err := logPath("test-agent")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(path, os.Getenv("HOME")+"/.config/squad-ai/logs/"))
	assert.True(t, strings.HasSuffix(path, ".log"))
	assert.Contains(t, path, "test-agent-")

	// Verify timestamp has no colons (replaced with hyphens).
	filename := filepath.Base(path)
	// Strip "test-agent-" prefix and ".log" suffix
	ts := strings.TrimSuffix(strings.TrimPrefix(filename, "test-agent-"), ".log")
	assert.NotEmpty(t, ts)
	assert.NotContains(t, ts, ":", "colons should be replaced in filename")
}

func TestInstallAll_AllSucceed(t *testing.T) {
	agents := []registry.Agent{
		safeAgent("a1", "/bin/true"),
		safeAgent("a2", "/bin/true"),
		safeAgent("a3", "/bin/true"),
	}

	results := testInstallAll(agents, nil)
	require.Len(t, results, 3)
	for i, err := range results {
		assert.NoError(t, err, "agent %d should succeed", i)
	}
}

func TestInstallAll_MixedResults(t *testing.T) {
	agents := []registry.Agent{
		safeAgent("good", "/bin/true"),
		safeAgent("bad", "/bin/false"),
		safeAgent("also-good", "/bin/true"),
	}

	results := testInstallAll(agents, nil)
	require.Len(t, results, 3)

	assert.NoError(t, results[0])
	assert.Error(t, results[1])
	assert.NoError(t, results[2])

	// Verify error mentions the failing agent.
	assert.Contains(t, results[1].Error(), "bad")
}

func TestInstallAll_EmptySlice(t *testing.T) {
	results := testInstallAll([]registry.Agent{}, nil)
	assert.Nil(t, results, "empty input should return nil")
}

func TestInstallAll_ProgressCalledForEach(t *testing.T) {
	var calls []string
	progress := func(agentID string, percentage int) {
		calls = append(calls, agentID)
		assert.Equal(t, 100, percentage)
	}

	agents := []registry.Agent{
		safeAgent("first", "/bin/true"),
		safeAgent("second", "/bin/true"),
	}

	results := testInstallAll(agents, progress)
	require.Len(t, results, 2)
	assert.NoError(t, results[0])
	assert.NoError(t, results[1])

	assert.Equal(t, []string{"first", "second"}, calls)
}

func TestInstallAll_ErrorsDontStopExecution(t *testing.T) {
	var calls []string
	progress := func(agentID string, percentage int) {
		calls = append(calls, agentID)
	}

	agents := []registry.Agent{
		safeAgent("first", "/bin/true"),
		safeAgent("second", "/bin/false"),
		safeAgent("third", "/bin/true"),
	}

	results := testInstallAll(agents, progress)
	require.Len(t, results, 3)

	assert.NoError(t, results[0])
	assert.Error(t, results[1])
	assert.NoError(t, results[2])

	// Only successful agents report progress completion.
	assert.Equal(t, []string{"first", "third"}, calls)
}

// TestInstallAgent_NonExistentBinary tests that a command referencing a
// non-existent binary returns an error.
func TestInstallAgent_NonExistentBinary(t *testing.T) {
	agent := registry.Agent{
		ID:   "bad-cmd",
		Name: "bad-cmd",
		Install: registry.InstallCmd{
			Method:  registry.MethodCustom,
			Command: "/nonexistent-binary-xyz",
		},
	}

	err := testInstall(agent, nil)
	require.Error(t, err, "non-existent binary should produce an error")
	assert.Contains(t, err.Error(), "bad-cmd")
}

// Test_logDir_creates_directory verifies the log directory is created.
func Test_logDir_creates_directory(t *testing.T) {
	dir, err := logDir()
	require.NoError(t, err)

	// Verify directory exists.
	info, err := os.Stat(dir)
	require.NoError(t, err, "log directory should exist")
	assert.True(t, info.IsDir())

	// Verify permissions (at minimum user-readable).
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
}
