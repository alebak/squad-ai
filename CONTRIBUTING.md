# Contributing to Squad AI

Thanks for your interest in contributing to Squad AI — a Go CLI tool that manages AI coding agents inside dev containers.

## Project Status

Squad AI is in early MVP development. The current workflow is streamlined for a small team. As the project grows, stricter automation (CI checks, label enforcement) will be introduced.

## Getting Started

### Prerequisites

- Go 1.24+
- Git

### Build and Run

```bash
git clone https://github.com/alebak/squad-ai.git
cd squad-ai
go build -o squad ./cmd/squad
./squad
```

## Development Workflow

### 1. Pick or create an issue

All changes start with an issue. If one doesn't exist, create it using the appropriate template:

- [Bug Report](https://github.com/alebak/squad-ai/issues/new?template=bug_report.yml)
- [Feature Request](https://github.com/alebak/squad-ai/issues/new?template=feature_request.yml)

Blank issues are disabled. Use [Discussions](https://github.com/alebak/squad-ai/discussions) for questions and ideas.

### 2. Branch naming

Branch names follow this convention:

```
<type>/<short-description>
```

| Prefix | Use for |
|--------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation only |
| `refactor/` | Code restructuring (no behavior change) |
| `chore/` | Maintenance, dependencies, tooling |
| `test/` | Adding or updating tests |

**Rules:** all lowercase, hyphens for separators, short and descriptive.

**Examples:** `feat/registry-cache`, `fix/tui-selection-bug`, `docs/api-reference`

### 3. Commit convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/).

```
<type>(<optional-scope>): <description>
```

**Allowed types:** `feat`, `fix`, `docs`, `refactor`, `chore`, `style`, `perf`, `test`, `build`, `ci`, `revert`

**Examples:**

```
feat(registry): add local cache with TTL
fix(installer): handle missing HOME env var
docs: add contributing guide
refactor(tui): extract agent list model
test(registry): add cache expiration tests
```

### 4. Open a pull request

- Reference the issue in the PR body: `Closes #N`
- Keep PRs focused — one logical change per PR
- Ensure tests pass: `go test ./...`
- Use the same Conventional Commits format for the PR title

## Code Standards

Squad AI follows the conventions documented in [AGENTS.md](AGENTS.md). Key points:

- Idiomatic Go following Effective Go and Go Code Review Comments
- Exported symbols must have godoc comments
- Handle every error — never use `_` to discard errors
- No `panic()` in library code
- Minimize external dependencies — stdlib is preferred
- Table-driven tests with testify

## Project Structure

```
cmd/squad/          Entry point
internal/cli/       Cobra command definitions
internal/tui/       Bubbletea models and views
internal/registry/  Remote registry fetch, cache, and parsing
internal/installer/ Agent installation and progress tracking
internal/config/    User configuration read/write
internal/runtime/   Runtime dependency detection
```

## Questions?

Use [GitHub Discussions](https://github.com/alebak/squad-ai/discussions) for questions, ideas, and general conversation.
