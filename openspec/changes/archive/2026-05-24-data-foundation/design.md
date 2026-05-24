# Design: Data Foundation — Registry + Config Models

## Technical Approach

Three greenfield artifacts: (1) `internal/registry/` — flat Go types mirroring JSON, HTTP fetch with cache, typed install method consts; (2) `internal/config/` — atomic file read/write via temp+rename, `$HOME/.config/squad-ai/config.json`; (3) `registry/agents.json` — 7 MVP agents matching specs. All stdlib — no new dependencies.

## Architecture Decisions

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Flat struct vs interface+strategy | Flat is 50 LOC vs 120 LOC. 5/7 agents use curl\|bash — no meaningful variation yet. | **Flat struct** with typed `InstallMethod` string const |
| Global HTTP client vs per-call | Global saves allocation but is mutable state. AGENTS.md forbids global mutable state. | **Create client per call** in `Fetch()` using `http.DefaultClient` as fallback |
| `$HOME/.config` vs `os.UserConfigDir()` | XDG gives `~/Library/Application Support` on macOS — PRD assumes `~/.config`. MVP targets Linux containers. | **Hardcode `$HOME/.config/squad-ai/`** — revisit at macOS support |
| Atomic write vs plain WriteFile | Atomic is 3 extra LOC, prevents partial writes on crash. Config stores user's agent selection — corruption costs UX. | **Atomic write** via `os.CreateTemp` + `os.Rename` |
| Package-level const vs var for registry URL | Const is simpler but prevents enterprise override (PRD §14 Fase 3). | **`var RegistryURL`** — override-able without code change |

## Data Flow

```
repo: registry/agents.json  ── HTTP GET ──►  ~/.config/squad-ai/registry.cache.json
         │                                              │
         │                                  json.Unmarshal ▼
         │                                  []Agent (in-memory)
         │                                            │
         │                  ┌─────────────────────────┤
         ▼                  ▼                         ▼
   source of truth    TUI reads agents      Installer looks up agent
   (committed)        for selection         by ID from cache

~/.config/squad-ai/config.json  ── Load/Save ──►  Config (in-memory)
                                                    │ selected_agents []
                                                    │ install_options
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `registry/agents.json` | Create | Source-of-truth with 7 MVP agents |
| `internal/registry/agent.go` | Create | `Agent`, `InstallCmd`, `RuntimeDep`, `Checksum`, `Registry` types + `InstallMethod` typed consts |
| `internal/registry/client.go` | Create | `Fetch()`, `LoadCache()`, `SaveCache()`, `IsStale()` |
| `internal/registry/registry_test.go` | Create | Fetch with httptest, cache round-trip, IsStale scenarios |
| `internal/config/config.go` | Create | `Config`, `InstallOptions` types. `Load()`, `Save()`, `DefaultConfig()`, `ConfigPath()` |
| `internal/config/config_test.go` | Create | Round-trip, missing file, malformed JSON, atomic write |
| `internal/registry/.gitkeep` | Delete | Replaced by real files |

## Interfaces / Contracts

```go
// internal/registry/agent.go

package registry

// InstallMethod identifies how an agent is installed.
type InstallMethod string

const (
    MethodCurlBash  InstallMethod = "curl_bash"
    MethodNpmInstall InstallMethod = "npm_install"
    MethodCustom    InstallMethod = "custom"
)

type InstallCmd struct {
    Method  InstallMethod `json:"method"`
    URL     string        `json:"url"`
    Command string        `json:"command"`
}

type RuntimeDep struct {
    Runtime    string `json:"runtime"`
    MinVersion string `json:"min_version,omitempty"`
}

type Checksum struct {
    SHA256     string `json:"sha256"`
    VerifiedAt string `json:"verified_at"`
}

type Agent struct {
    ID           string       `json:"id"`
    Name         string       `json:"name"`
    Description  string       `json:"description"`
    Version      string       `json:"version"`
    DetectCmd    string       `json:"detect_command"`
    Install      InstallCmd   `json:"install"`
    Dependencies []RuntimeDep `json:"dependencies"`
    Checksum     *Checksum    `json:"checksum,omitempty"`
    Tags         []string     `json:"tags"`
    AddedAt      string       `json:"added_at"`
}

type Registry struct {
    Agents    []Agent `json:"agents"`
    UpdatedAt string  `json:"updated_at"`
}

// RegistryURL is the remote registry URL. Package-level var for future override.
var RegistryURL = "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json"

// Fetch downloads and parses the registry from url via HTTP GET.
// Uses context for cancellation. Returns wrapped error on non-200 or network failure.
func Fetch(ctx context.Context, url string) (*Registry, error)

// LoadCache reads and parses a Registry from a local JSON file path.
func LoadCache(path string) (*Registry, error)

// SaveCache writes a Registry to a local JSON file path.
func SaveCache(path string, r *Registry) error

// IsStale returns true if the cache file is older than maxAge.
// Returns false if the file does not exist.
func IsStale(path string, maxAge time.Duration) bool
```

```go
// internal/config/config.go

package config

type InstallOptions struct {
    Silent     bool `json:"silent"`
    PreferPnpm bool `json:"prefer_pnpm"`
}

type Config struct {
    SelectedAgents    []string       `json:"selected_agents"`
    RegistryLastCheck string         `json:"registry_last_checked"`
    RegistryKnown     []string       `json:"registry_known_agents"`
    InstallOptions    InstallOptions `json:"install_options"`
}

// DefaultConfig returns sensible defaults: Silent=true, PreferPnpm=true.
func DefaultConfig() *Config

// ConfigPath returns $HOME/.config/squad-ai/config.json.
// Creates parent directories with 0755 if missing. Errors if HOME is unset.
func ConfigPath() (string, error)

// Load reads config from path. Returns DefaultConfig if file missing.
// Returns error on malformed JSON or read failure (wrapped).
func Load(path string) (*Config, error)

// Save writes config atomically: temp file + os.Rename, 0644 permissions.
// Creates parent directories if missing.
func Save(path string, cfg *Config) error
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Agent JSON round-trip, `omitempty` behavior | Marshal/unmarshal in-memory, assert field equality |
| Unit | Registry file parsing (all 7 agents) | Parse `testdata/agents.json` or inline JSON |
| Unit | Fetch success, cancellation, non-200, network error | `httptest.NewServer` with canned responses |
| Unit | Cache round-trip (Save → Load identical) | Temp dir, compare parsed agents |
| Unit | IsStale (fresh, stale, missing) | Set file mtime, verify boolean |
| Unit | Config round-trip (marshal/unmarshal match) | In-memory, field-by-field assert |
| Unit | DefaultConfig on missing file | Remove file, call Load, assert defaults |
| Unit | Error on malformed JSON | Write garbage, call Load, assert error |
| Unit | Atomic write (temp file cleaned) | Save then verify final path exists, no `.tmp` lingering |
| Unit | ConfigPath directory creation | Set HOME, call ConfigPath, assert dir exists |

## Migration / Rollout

No migration required. Both artifacts are greenfield. `registry/agents.json` is version-controlled — future agents are added via PR. Config is ephemeral (recreated on next `squad run`).

## Open Questions

None. All decisions resolved in exploration and specs.
