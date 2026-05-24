## Exploration: Data Foundation — Registry + Config Models

### Current State

Project is bootstrapped with Cobra CLI, `squad version` working, and 6 empty internal packages (`registry`, `config`, `tui`, `installer`, `runtime`, `cli`). No data types, no registry, no config. The Go module at `github.com/alebak/squad-ai` has cobra and testify as dependencies only.

The PRD defines the full spec in sections 6.2–6.4:
- **Registry**: JSON file at `registry/agents.json` in the repo, fetched from `raw.githubusercontent.com`, cached to `~/.config/squad-ai/registry.cache.json`
- **Config**: User manifest at `~/.config/squad-ai/config.json` with `selected_agents`, `registry_last_checked`, etc.
- **Agent fields**: id, name, description, version, detect_command, install (method+url+command), dependencies, checksum, tags, added_at

7 agents confirmed for MVP (note: `aider` from PRD §13 is deferred):
1. Claude Code — curl | bash, no runtime
2. OpenCode — curl | bash, no runtime
3. Pi — curl | sh, no runtime
4. Codex — npm i -g @openai/codex, runtime: node
5. Antigravity CLI — curl | bash, no runtime
6. Gemini CLI — npm install -g @google/gemini-cli, runtime: node
7. Gentle-AI — curl | bash, no runtime

### Affected Areas

- `registry/agents.json` — **New file**. Source-of-truth registry in the repo
- `internal/registry/` — **New package**. Types, fetch, cache, parse
- `internal/registry/agent.go` — Agent struct and related types
- `internal/registry/client.go` — HTTP fetch + cache read/write logic
- `internal/registry/registry_test.go` — Tests for parsing and caching
- `internal/config/` — **New package**. User config types + read/write
- `internal/config/config.go` — Config struct, Load(), Save()
- `internal/config/config_test.go` — Tests for read/write round-trip
- `go.mod` — **Updated** (no new deps needed if using stdlib for HTTP/JSON/XDG)

### Approaches

#### Registry File — 3 Approaches

1. **PRD spec: repo file + fetch + cache** — `registry/agents.json` in repo, fetched on first run, cached locally with 24h TTL
   - Pros: Exact PRD match, agents can be added via PR without binary rebuild, offline-capable via cache
   - Cons: Requires HTTP on first run, cache management logic, versioning concerns (cache may be stale)
   - Effort: Medium (~120 LOC)

2. **Go embed: ship registry inside the binary** — Use `//go:embed registry/agents.json` to bake registry into the binary
   - Pros: No network dependency for initial data, simpler implementation, deterministic (binary == registry)
   - Cons: Updating agents requires new binary release, PRD says remote URL, defeats the "discover new agents" flow
   - Effort: Low (~60 LOC)

3. **Hybrid: embedded baseline + remote updates** — Ship minimal registry in binary, attempt remote fetch on each run, merge updates
   - Pros: Works offline for known agents, still discovers new ones
   - Cons: Merge logic complexity, dual source-of-truth questions, over-engineering for MVP
   - Effort: High (~200 LOC)

#### Agent Model — 2 Approaches

1. **Flat struct, one-to-one with JSON** — Direct Go struct mirroring the PRD JSON. Use `InstallMethod` as typed string const. `Dependency` as a struct with `Runtime` and optional `MinVersion`.
   - Pros: Simple, zero magic, easy to test, natural JSON round-trip
   - Cons: Less type safety on install method values (can be mitigated with validation)
   - Effort: Low (~50 LOC)

2. **Interface + strategy pattern for install methods** — `Installer` interface with `CurlBashInstaller`, `NpmInstaller`, etc. Each agent holds an `Installer` rather than a raw command string.
   - Pros: Cleaner for execution phase, each install method can have its own logic/validation
   - Cons: Over-engineering for MVP — 5 of 7 agents use curl|bash. Adds abstraction layer before we know what execution patterns emerge. The raw command string in the registry is sufficient.
   - Effort: Medium (~120 LOC)

#### Config Model — 2 Approaches

1. **Standard file read/write with encoding/json** — `os.ReadFile` + `json.Unmarshal` on load, `json.Marshal` + `os.WriteFile` on save. XDG path via `os.UserConfigDir()`.
   - Pros: Zero new dependencies, dead simple, 100% stdlib
   - Cons: No atomic writes (can corrupt file if interrupted mid-write — acceptable for CLI tool)
   - Effort: Low (~80 LOC)

2. **Buffered write with atomic rename** — Same as above but write to a temp file, then `os.Rename` for atomicity. Optionally add a mutex.
   - Pros: No partial writes, clean
   - Cons: Slightly more code, but still stdlib. Worth doing for safety — partial config.json on crash would lose user's agent selection.
   - Effort: Low (~100 LOC)

### Recommendation

**Use the hybrid of:**

- **Registry**: Approach 1 (PRD spec: repo file + fetch + cache). It's what the PRD says, it enables the "discover new agents" flow (PRD §7.2 step 5-6), and the cache handles offline scenarios. The HTTP dependency is stdlib `net/http` — no new deps.
- **Agent Model**: Approach 1 (flat struct). Keep it simple. The install method can be a typed string const (`const MethodCurlBash = "curl_bash"`) for now. The strategy pattern comes later when install methods diverge meaningfully.
- **Config**: Approach 2 (atomic write). The cost is minimal (3 extra LOC for temp file + rename) and protects against corrupted config on crash/interrupt — important since config stores the user's agent selection.

**Connection between registry and config**:
```
registry/agents.json (repo)
        │
        ▼  HTTP GET (first run or stale)
~/.config/squad-ai/registry.cache.json
        │
        ▼  json.Unmarshal
[]Agent (in-memory, internal/registry)
        │
        ├──► TUI reads []Agent for selection UI
        │
        └──► selected_agents []string ──► config.json
                                              │
                                              ▼  json.Unmarshal
                                         Config (in-memory, internal/config)
                                              │
                                              ▼  installer looks up Agent by ID
                                         Agent details from registry cache
```

### Proposed Go Types

```go
// internal/registry/agent.go

package registry

// Agent represents a single AI coding agent entry in the registry.
type Agent struct {
    ID           string         `json:"id"`
    Name         string         `json:"name"`
    Description  string         `json:"description"`
    Version      string         `json:"version"`       // "latest" or semver
    DetectCmd    string         `json:"detect_command"`
    Install      InstallCmd     `json:"install"`
    Dependencies []RuntimeDep   `json:"dependencies"`
    Checksum     *Checksum      `json:"checksum,omitempty"`
    Tags         []string       `json:"tags"`
    AddedAt      string         `json:"added_at"`       // RFC3339
}

// InstallCmd describes how to install an agent.
type InstallCmd struct {
    Method  string `json:"method"`  // "curl_bash", "npm_install", "go_install", "binary_download", "custom"
    URL     string `json:"url"`
    Command string `json:"command"` // full install command string
}

// RuntimeDep declares a runtime dependency with optional minimum version.
type RuntimeDep struct {
    Runtime   string `json:"runtime"`             // "node", "go", "python", "none"
    MinVersion string `json:"min_version,omitempty"` // semver constraint
}

// Checksum holds the SHA-256 hash for integrity verification.
type Checksum struct {
    SHA256     string `json:"sha256"`
    VerifiedAt string `json:"verified_at"` // RFC3339
}

// Registry wraps a list of agents fetched from the remote.
type Registry struct {
    Agents    []Agent `json:"agents"`
    UpdatedAt string  `json:"updated_at"` // when this registry was published
}
```

```go
// internal/config/config.go

package config

// Config represents the user's agent selection manifest.
type Config struct {
    SelectedAgents    []string       `json:"selected_agents"`
    RegistryLastCheck string         `json:"registry_last_checked"`  // RFC3339
    RegistryKnown     []string       `json:"registry_known_agents"`
    InstallOptions    InstallOptions `json:"install_options"`
}

// InstallOptions holds global installation preferences.
type InstallOptions struct {
    Silent     bool `json:"silent"`
    PreferPnpm bool `json:"prefer_pnpm"`
}
```

### Risks

- **Checksum availability**: Writing real SHA-256 checksums for `curl | bash` scripts means someone must manually compute them. For MVP, the checksum field MAY be empty, and verification can be a best-effort warning. The PRD says MVP includes the field but Fase 2 locks it down.
- **Registry URL hardcoding**: The raw GitHub URL is hardcoded. For MVP this is fine, but should be configurable for future enterprise registries. Mitigation: use a package-level `var RegistryURL = "..."` so it can be overridden later.
- **XDG path differences**: `os.UserConfigDir()` returns different paths on Linux (`~/.config`), macOS (`~/Library/Application Support`), and Windows. The PRD assumes `~/.config/squad-ai/`. On macOS this becomes `~/Library/Application Support/squad-ai/`. This needs explicit acceptance or a decision to override for consistency.
- **No existing registry/agents.json**: Must create the initial file with 7 agents. This is the first "content" file in the repo — ensure the JSON is valid and matches the Go types exactly to avoid parsing errors.

### Ready for Proposal

Yes — the scope is clear and well-bounded. Proceed to `sdd-propose` with change name `data-foundation`.

**What to tell the user**: "The data foundation change covers 3 packages (`internal/registry`, `internal/config`, and the `registry/agents.json` file). It establishes the Go types, fetches and caches the registry via HTTP, reads/writes config with atomic file operations, and creates the initial agents.json with all 7 MVP agents. No TUI, no installer code — just data plumbing. Effort is ~300 LOC across ~10 files, within the 400-line budget."
