package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Real-system detection tests.

func TestDetectGo_RealBinary(t *testing.T) {
	info := DetectGo()
	assert.True(t, info.Installed, "go should be installed in the dev container")
	assert.Equal(t, "go", info.Name)
	assert.NotEmpty(t, info.Version, "go version should be non-empty")
	assert.Contains(t, info.RawOutput, "go version go")
	assert.NoError(t, info.Err)
}

func TestDetectPython_RealBinary(t *testing.T) {
	info := DetectPython()
	assert.True(t, info.Installed, "python3 should be installed in the dev container")
	assert.Equal(t, "python", info.Name)
	assert.NotEmpty(t, info.Version, "python version should be non-empty")
	assert.Contains(t, info.RawOutput, "Python ")
	assert.NoError(t, info.Err)
}

func TestDetectNode_NotInstalled(t *testing.T) {
	info := DetectNode()
	assert.False(t, info.Installed, "node is not expected in this dev container")
	assert.Equal(t, "node", info.Name)
	assert.Empty(t, info.Version)
	assert.NoError(t, info.Err)
}

// Table-driven parser tests.

func TestParseNodeVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "standard", input: "v22.3.0\n", want: "22.3.0"},
		{name: "no newline", input: "v18.12.1", want: "18.12.1"},
		{name: "old version", input: "v12.22.12\n", want: "12.22.12"},
		{name: "garbage", input: "not a version\n", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNodeVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "standard", input: "go version go1.24.0 linux/amd64\n", want: "1.24.0"},
		{name: "old version", input: "go version go1.16.5 darwin/arm64\n", want: "1.16.5"},
		{name: "new version", input: "go version go1.26.3 linux/amd64\n", want: "1.26.3"},
		{name: "no go prefix", input: "go version 1.24.0 linux/amd64\n", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGoVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParsePythonVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "standard", input: "Python 3.12.1\n", want: "3.12.1"},
		{name: "no newline", input: "Python 3.11.5", want: "3.11.5"},
		{name: "old version", input: "Python 2.7.18\n", want: "2.7.18"},
		{name: "garbage", input: "garbage output\n", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePythonVersion(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Table-driven compareVersions tests (5+ pairs).

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "a greater than b", a: "22.3.0", b: "18.0.0", want: 1},
		{name: "a less than b", a: "16.0.0", b: "18.0.0", want: -1},
		{name: "equal versions", a: "22.3.0", b: "22.3.0", want: 0},
		{name: "same major, minor greater", a: "22.4.0", b: "22.3.0", want: 1},
		{name: "same major, minor less", a: "22.2.0", b: "22.3.0", want: -1},
		{name: "same major.minor, patch greater", a: "22.3.5", b: "22.3.0", want: 1},
		{name: "same major.minor, patch less", a: "22.3.0", b: "22.3.5", want: -1},
		{name: "malformed a returns 0", a: "abc", b: "18.0.0", want: 0},
		{name: "malformed b returns 0", a: "18.0.0", b: "abc", want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Table-driven IsCompatible tests.

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		name       string
		info       RuntimeInfo
		minVersion string
		want       bool
	}{
		{
			name:       "version meets min",
			info:       RuntimeInfo{Installed: true, Version: "22.3.0"},
			minVersion: "18.0.0",
			want:       true,
		},
		{
			name:       "version below min",
			info:       RuntimeInfo{Installed: true, Version: "16.0.0"},
			minVersion: "18.0.0",
			want:       false,
		},
		{
			name:       "version equals min",
			info:       RuntimeInfo{Installed: true, Version: "18.0.0"},
			minVersion: "18.0.0",
			want:       true,
		},
		{
			name:       "not installed",
			info:       RuntimeInfo{Installed: false},
			minVersion: "18.0.0",
			want:       false,
		},
		{
			name:       "malformed installed version",
			info:       RuntimeInfo{Installed: true, Version: "abc"},
			minVersion: "18.0.0",
			want:       false,
		},
		{
			name:       "malformed min version",
			info:       RuntimeInfo{Installed: true, Version: "22.0.0"},
			minVersion: "abc",
			want:       false,
		},
		{
			name:       "empty version string",
			info:       RuntimeInfo{Installed: true, Version: ""},
			minVersion: "18.0.0",
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCompatible(tt.info, tt.minVersion)
			assert.Equal(t, tt.want, got)
		})
	}
}
