package installer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// safeAgent returns a registry.Agent whose install command is a safe
// command (no real agent installation). Use this in all tests to avoid
// executing actual install scripts.
func safeAgent(id, cmd string) registry.Agent {
	return registry.Agent{
		ID:   id,
		Name: id,
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
	err := InstallAgent(agent, nil)
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
	err := InstallAgent(agent, progress)
	assert.NoError(t, err)
	assert.True(t, progressCalled, "progress callback should be called on success")
}

func TestInstallAgent_Failure(t *testing.T) {
	agent := safeAgent("fail-agent", "/bin/false")
	err := InstallAgent(agent, nil)
	require.Error(t, err)

	// Error should mention exit code 1
	assert.Contains(t, err.Error(), "exited with code 1")
	assert.Contains(t, err.Error(), "fail-agent")
}

func TestInstallAgent_OutputCapturedToLog(t *testing.T) {
	agent := safeAgent("echo-test", "echo hello from install")
	err := InstallAgent(agent, nil)
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
			data, err := os.ReadFile(filepath.Join(logDir, e.Name()))
			require.NoError(t, err)
			assert.Contains(t, string(data), "hello from install")
			break
		}
	}
	assert.True(t, found, "log file for echo-test should exist")
}

func TestInstallAgent_NoChecksumWarns(t *testing.T) {
	// Agent with nil checksum should still install successfully.
	agent := safeAgent("no-cksum", "/bin/true")
	err := InstallAgent(agent, nil)
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

	results := InstallAll(agents, nil)
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

	results := InstallAll(agents, nil)
	require.Len(t, results, 3)

	assert.NoError(t, results[0])
	assert.Error(t, results[1])
	assert.NoError(t, results[2])

	// Verify error mentions the failing agent.
	assert.Contains(t, results[1].Error(), "bad")
}

func TestInstallAll_EmptySlice(t *testing.T) {
	results := InstallAll([]registry.Agent{}, nil)
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

	results := InstallAll(agents, progress)
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

	results := InstallAll(agents, progress)
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

	err := InstallAgent(agent, nil)
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


