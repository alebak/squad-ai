## Exploration: Agent Uninstall Support

### Current State

- `squad remove <id>` only removes an agent from `selected_agents` in config — it does NOT uninstall the binary from the system.
- The remove command explicitly states: *"Note: the agent is still installed on your system if present."*
- The PRD (§8) confirms this scope: `squad remove <id>` "Remueve un agente del manifiesto (no lo desinstala del sistema)."
- The `registry.Agent` struct has `Install.InstallCmd` but no uninstall counterpart.
- The `installer` package has `InstallAgent()` but no `UninstallAgent()`.
- The `validateCommand` function validates install methods (`curl_bash`, `npm_install`, `custom`) but has no uninstall validation.
- Current agents in the registry:
  - `npm_install` agents (2): codex, gemini-cli — have a clean `npm uninstall -g` path
  - `curl_bash` agents (5): claude-code, opencode, pi, antigravity-cli, gentle-ai — typically no standardized uninstall; best-effort binary deletion only

### Affected Areas

- `internal/registry/agent.go` — Add `UninstallCmd` field to `InstallCmd` (or new struct) and optional `uninstall` method constant
- `internal/installer/install.go` — Add `UninstallAgent` function following `InstallAgent` pattern
- `internal/installer/installer.go` — Potentially add `FindAgentBinary` helper (to resolve `detect_command` to full binary path)
- `internal/cli/remove.go` — Wire uninstall into the remove flow (behind `--uninstall` flag or default behavior)
- `internal/cli/remove_test.go` — Update tests for new behavior
- `registry/agents.json` — Add uninstall commands for known agents
- `openspec/specs/remove/spec.md` — Update spec to cover uninstall requirement

### Approaches

1. **Registry-driven uninstall (explicit `uninstall` field)** — Add optional `uninstall.command` and `uninstall.method` to each agent's registry entry. `UninstallAgent` reads and executes it. If no `uninstall.command` is set, fall back to method-derived behavior.
   - Pros: Explicit per-agent control; handles edge cases (e.g., config files, cache dirs); follows existing `install` pattern exactly; cleanest architecture
   - Cons: Requires manual maintenance of uninstall commands in registry; older agents won't have it until someone adds it
   - Effort: Medium

2. **Method-derived automatic uninstall** — Derive uninstall command from the install method automatically: `npm uninstall -g <package>` for `npm_install`, `rm -f $(which <detect_command>)` for `curl_bash`, etc. No registry changes needed.
   - Pros: Zero registry maintenance; works for all present and future agents automatically
   - Cons: Fragile — `rm -f $(which cmd)` leaves config/cache behind; extracting npm package name from install command is heuristic; no way to override per-agent
   - Effort: Low

3. **Hybrid: registry-driven with method-derived fallback** — Add optional `uninstall` field to registry. If present, use it. If absent, derive best-effort uninstall from method + `detect_command`. This lets registry curators override the default for agents that need cleanup beyond binary deletion.
   - Pros: Flexibility + automation; no breaking changes for agents without uninstall data; cleanest UX
   - Cons: Slightly more complex implementation than either pure approach; need to define the fallback logic carefully per method
   - Effort: Medium

4. **Minimal: `--uninstall` flag on `squad remove` with manual uninstall only** — Add a `--uninstall` flag to `squad remove`. When set, it prompts "uninstall not available for this agent; manually delete?" and/or just removes from config + prints instructions for manual uninstall.
   - Pros: Simplest change; minimal risk; preserves backward compatibility
   - Cons: Doesn't actually uninstall anything; user still has to do it manually; doesn't solve the problem
   - Effort: Low

### Recommendation

**Approach 3 (Hybrid: registry-driven with method-derived fallback)** — This is the right balance:

- Add `uninstall` field to `InstallCmd` in the registry JSON model:
  ```go
  type InstallCmd struct {
      Method         InstallMethod `json:"method"`
      URL            string        `json:"url"`
      Command        string        `json:"command"`
      NonInteractive bool          `json:"non_interactive"`
      Uninstall      *UninstallCmd `json:"uninstall,omitempty"`
  }

  type UninstallCmd struct {
      Command string `json:"command"`
  }
  ```
- Implement `UninstallAgent(agent registry.Agent, progress ProgressFn)` in the installer package:
  1. If `agent.Install.Uninstall` is set and non-empty, execute it via `sh -c` (same pattern as install)
  2. If not set, derive fallback:
     - `npm_install` → extract package name from install command, run `npm uninstall -g <package>`
     - `curl_bash` → delete binary at `$(which <detect_command>)` via `rm -f`
     - `custom` → no fallback; return "no uninstall method defined"
- Wire into `squad remove` with an explicit `--uninstall` flag (default false for backward compat)
- Update `registry/agents.json` with explicit uninstall commands for known agents where possible (codex, gemini-cli via npm; curl_bash agents get their fallback)
- Update the remove spec to add the uninstall requirement

This keeps the PRD's stated behavior as default (config-only removal) and adds the actual uninstall opt-in.

### Risks

- **Destructive operation**: `rm -f $(which <detect>)` is irreversible. A mistyped `detect_command` in the registry could delete the wrong binary. Mitigation: show what will be removed and confirm with user before executing (at least the first time, or when `--uninstall` is used without `--force`).
- **npm uninstall side effects**: `npm uninstall -g` may fail if the package wasn't installed via npm, or if npm is not available. Mitigation: check runtime before attempting (follow existing install pattern with `isRuntimeMet`).
- **False positives on detect_command**: If two unrelated tools share the same binary name, `which` returns only one. Mitigation: document that uninstall targets the binary matched by `detect_command`; if the user changed it, they must manually uninstall.
- **Breaks existing remove behavior**: Current `squad remove` is purely config. Adding uninstall changes the mental model. Mitigation: use `--uninstall` flag, keep `squad remove` as config-only by default.

### Ready for Proposal

Yes. The hybrid approach is well-defined, the registry schema change is backward-compatible, and the implementation follows existing patterns (InstallAgent, runAndLog, sh -c execution with logging). The orchestrator should move to `sdd-spec` next.
