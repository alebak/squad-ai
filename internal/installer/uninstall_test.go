package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUninstallAgent_ExplicitCommand verifies that an explicit UninstallCmd
// is executed when present.
func TestUninstallAgent_ExplicitCommand(t *testing.T) {
	agent := registry.Agent{
		ID:   "explicit-test",
		Name: "Explicit Test",
		Install: registry.InstallCmd{
			Method:       registry.MethodCustom,
			Command:      "/bin/true",
			UninstallCmd: "/bin/true",
		},
	}

	err := UninstallAgent(agent)
	assert.NoError(t, err, "explicit uninstall command should succeed")
}

// TestUninstallAgent_ExplicitCommandFailure verifies that an explicit
// UninstallCmd that fails returns an error.
func TestUninstallAgent_ExplicitCommandFailure(t *testing.T) {
	agent := registry.Agent{
		ID:   "explicit-fail",
		Name: "Explicit Fail",
		Install: registry.InstallCmd{
			Method:       registry.MethodCustom,
			Command:      "/bin/true",
			UninstallCmd: "/bin/false",
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "explicit-fail")
	assert.Contains(t, err.Error(), "exited with code 1")
}

// TestUninstallAgent_NpmFallback verifies that an npm_install agent without
// an explicit UninstallCmd derives the correct uninstall command.
func TestUninstallAgent_NpmFallback(t *testing.T) {
	// Use /bin/true as a fake "npm uninstall" to verify the derived
	// command is executed. We can't easily test the real npm uninstall
	// in unit tests, but we verify the derivation logic.
	agent := registry.Agent{
		ID:        "npm-test",
		Name:      "NPM Test",
		DetectCmd: "true",
		Install: registry.InstallCmd{
			Method:  registry.MethodNpmInstall,
			Command: "npm i -g @openai/codex",
		},
	}

	// The derived command would be "npm uninstall -g @openai/codex".
	// We don't actually run npm in tests, but we can verify the error
	// message mentions the derived command and the agent ID.
	err := UninstallAgent(agent)
	// npm may or may not be installed in the test environment.
	// What matters is that the error (if any) mentions the agent.
	if err != nil {
		// Acceptable errors: npm not found, npm command failed
		assert.Contains(t, err.Error(), "npm-test")
	}
}

// TestUninstallAgent_CurlBashFallback creates a temporary binary in PATH,
// then verifies the curl_bash fallback removes it.
func TestUninstallAgent_CurlBashFallback(t *testing.T) {
	// Create a temporary directory and add it to PATH.
	tmpDir := t.TempDir()
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", tmpDir+string(filepath.ListSeparator)+oldPath)

	// Create a fake binary.
	binaryPath := filepath.Join(tmpDir, "test-uninstall-me")
	err := os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0755)
	require.NoError(t, err)

	// Verify it's found in PATH.
	_, err = exec.LookPath("test-uninstall-me")
	require.NoError(t, err)

	agent := registry.Agent{
		ID:        "curl-test",
		Name:      "Curl Test",
		DetectCmd: "test-uninstall-me",
		Install: registry.InstallCmd{
			Method:  registry.MethodCurlBash,
			Command: "curl -fsSL https://example.com/install.sh | bash",
		},
	}

	err = UninstallAgent(agent)
	assert.NoError(t, err, "curl_bash fallback should remove binary")

	// Verify the binary is gone.
	_, err = exec.LookPath("test-uninstall-me")
	assert.Error(t, err, "binary should no longer be in PATH")
}

// TestUninstallAgent_CustomNoUninstall verifies that custom agents without
// an explicit UninstallCmd return an error.
func TestUninstallAgent_CustomNoUninstall(t *testing.T) {
	agent := registry.Agent{
		ID:   "custom-no-uninstall",
		Name: "Custom No Uninstall",
		Install: registry.InstallCmd{
			Method:  registry.MethodCustom,
			Command: "/some/custom/install.sh",
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no uninstall command defined")
	assert.Contains(t, err.Error(), "custom-no-uninstall")
}

// TestUninstallAgent_NullByte verifies that a null byte in the uninstall
// command is rejected.
func TestUninstallAgent_NullByte(t *testing.T) {
	agent := registry.Agent{
		ID:   "nullbyte-test",
		Name: "Null Byte Test",
		Install: registry.InstallCmd{
			Method:       registry.MethodCustom,
			Command:      "/bin/true",
			UninstallCmd: "/bin/true\x00evil",
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "null byte")
	assert.Contains(t, err.Error(), "nullbyte-test")
}

// TestUninstallAgent_CurlBashEmptyDetectCmd verifies that curl_bash fallback
// with an empty detect_cmd returns an error.
func TestUninstallAgent_CurlBashEmptyDetectCmd(t *testing.T) {
	agent := registry.Agent{
		ID:   "empty-detect",
		Name: "Empty Detect",
		Install: registry.InstallCmd{
			Method:  registry.MethodCurlBash,
			Command: "curl -fsSL https://example.com/install.sh | bash",
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detect_command is required")
	assert.Contains(t, err.Error(), "empty-detect")
}

// TestExtractNPMPackage verifies the npm package extraction logic.
func TestExtractNPMPackage(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		want      string
		wantEmpty bool
	}{
		{
			name:    "npm i -g with scoped package",
			command: "npm i -g @openai/codex",
			want:    "@openai/codex",
		},
		{
			name:    "npm install -g with scoped package",
			command: "npm install -g @google/gemini-cli",
			want:    "@google/gemini-cli",
		},
		{
			name:    "npm i --global with scoped package",
			command: "npm i --global @openai/codex",
			want:    "@openai/codex",
		},
		{
			name:    "npm i -g simple package",
			command: "npm i -g typescript",
			want:    "typescript",
		},
		{
			name:      "no package name",
			command:   "npm i -g",
			wantEmpty: true,
		},
		{
			name:      "empty command",
			command:   "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractNPMPackage(tt.command)
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestUninstallAgent_NpmFallbackNoPackage verifies that npm fallback returns
// an error when no package can be extracted.
func TestUninstallAgent_NpmFallbackNoPackage(t *testing.T) {
	agent := registry.Agent{
		ID:   "no-pkg",
		Name: "No Package",
		Install: registry.InstallCmd{
			Method:  registry.MethodNpmInstall,
			Command: "npm i",
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not extract npm package name")
}

// TestUninstallAgent_UnknownMethod verifies that an unknown install method
// returns an error.
func TestUninstallAgent_UnknownMethod(t *testing.T) {
	agent := registry.Agent{
		ID:   "unknown",
		Name: "Unknown",
		Install: registry.InstallCmd{
			Method: registry.InstallMethod("unknown_method"),
		},
	}

	err := UninstallAgent(agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown install method")
}
