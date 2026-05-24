## Exploration: Agent Installation with Progress Tracking

### Current State

The `internal/installer/` package currently has only `DetectAgent` and `DetectAll` — checking if an agent binary exists in PATH. There is no installation capability. The `registry.Agent` type has an `Install` field of type `InstallCmd` with `Method`, `URL`, and `Command` fields ready for use.

### Affected Areas

- `internal/installer/install.go` — NEW file: InstallAgent, InstallAll, ProgressFn type, log path generation
- `internal/installer/install_test.go` — NEW file: tests using /bin/true, /bin/false, echo
- `internal/installer/installer.go` — No changes (has DetectAgent/DetectAll)
- `internal/config/config.go` — Existing pattern for config path generation (reference for log dir)

### Approaches

1. **Single-file approach** — One `install.go` with all installation logic
   - Pros: Simple, follows existing pattern (installer.go has DetectAgent+DetectAll), easy to review
   - Cons: None for this scope
   - Effort: Low

2. **Split into files** — install.go (core) + progress.go (types) + log.go (logging)
   - Pros: Separation of concerns
   - Cons: Premature for ~120 lines, adds cognitive overhead
   - Effort: Low

### Recommendation

Approach 1: Single `install.go` file. The functionality is cohesive (install agent, track progress, capture logs) and fits naturally in one file at this stage. Follow the exact pattern of the existing `installer.go`.

### Risks

- Shell commands via `exec.Command("sh", "-c", command)` are inherently risky — but this is by design (agents use curl|bash scripts)
- Log directory creation might fail on read-only filesystems (containers) — handled gracefully with warning
- Tests must never execute real agent install commands — use echo/true/false only

### Ready for Proposal

Yes. The requirements are clear from PRD §10 and §11. Single file addition to `internal/installer/`, well-scoped.
