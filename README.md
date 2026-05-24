# Squad AI

**Manage your AI coding agent squad inside dev containers.**

[![Go 1.24+](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey)](https://github.com/alebak/squad-ai)

---

## What It Does

Squad AI is a CLI that installs and manages AI coding agents (Claude Code, OpenCode, Pi, Codex, and more) inside dev containers. It replaces ad-hoc `post-create.sh` scripts with a single tool.

**Before**: copying `curl | bash` snippets between projects, keeping them in sync manually.  
**After**: `squad add` → pick your agents → done. One command in `postCreateCommand`.

## Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
```

This downloads the latest binary to `~/.local/bin/squad`. Add it to your PATH if needed:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Usage

```bash
# First run: pick your agents interactively
squad add

# Install your selected agents silently
squad install

# Install specific agents (no TUI)
squad install --agents claude-code,opencode,gentle-ai

# See what's installed and available
squad list

# Update the agent registry
squad update
```

## Dev Container Integration

Add to your `.devcontainer/devcontainer.json`:

```json
{
  "postCreateCommand": "squad install"
}
```

Or use a `post-create.sh` for first-time setup:

```bash
#!/bin/bash
set -e

if ! command -v squad &>/dev/null; then
  curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
  export PATH="$HOME/.local/bin:$PATH"
fi

squad install
```

Config persists between container rebuilds when `$HOME` is mounted as a volume.

## Supported Agents

| Agent | Install Method | Runtime |
|-------|---------------|---------|
| Claude Code | curl \| bash | none |
| OpenCode | curl \| bash | none |
| Pi | curl \| bash | none |
| Codex CLI | npm | Node.js |
| Antigravity CLI | curl \| bash | none |
| Gemini CLI | npm | Node.js |
| Gentle AI | curl \| bash | none |

New agents are added via PR to `registry/agents.json`. Squad AI notifies you when new agents become available.

## How It Works

1. **Registry** — a `registry/agents.json` file in this repo lists all available agents with install commands, runtime dependencies, and SHA-256 checksums.
2. **Local cache** — squad fetches and caches the registry locally (`~/.config/squad-ai/registry.cache.json`).
3. **Detection** — on `squad list`, it checks which agents are already installed and which runtimes are missing.
4. **Installation** — `squad install` runs the install commands, captures logs, verifies checksums, and saves your selection.
5. **TUI** — `squad add` opens a terminal UI where you browse, select, and install agents interactively.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow. Quick overview:

- **Issue-first**: open an issue before a PR
- **Conventional Commits**: `feat(scope):`, `fix(scope):`, `docs:`, etc.
- **Branch naming**: `feat/`, `fix/`, `docs/`, `chore/`
- **Code standards**: idiomatic Go, stdlib first, godoc on exports

Questions? Use [Discussions](https://github.com/alebak/squad-ai/discussions), not issues.

---

## Español / Spanish

Squad AI es un CLI en Go que gestiona la instalación de coding agents dentro de dev containers. Reemplaza los scripts `post-create.sh` manuales por una herramienta unificada.

### Instalación rápida

```bash
curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
```

### Uso

| Comando | Descripción |
|---------|-------------|
| `squad add` | TUI interactivo para elegir agentes |
| `squad install` | Instala los agentes seleccionados |
| `squad install --agents a,b,c` | Instala agentes específicos sin TUI |
| `squad list` | Lista agentes: instalados, disponibles, bloqueados |
| `squad info <id>` | Detalle de un agente |
| `squad remove <id>` | Quita un agente de tu configuración |
| `squad update` | Actualiza el registro de agentes |
| `squad version` | Versión de Squad AI |

### ¿Cómo funciona?

1. **Registro** — `registry/agents.json` lista los agentes disponibles con comandos de instalación, dependencias y checksums.
2. **Cache local** — squad descarga y cachea el registro en `~/.config/squad-ai/`.
3. **Detección** — verifica qué agentes ya están instalados y qué runtimes faltan.
4. **Instalación** — ejecuta los comandos, captura logs, verifica checksums y guarda tu selección.
5. **TUI** — `squad add` abre una interfaz interactiva para explorar y seleccionar agentes.

### Contribuir

Ver [CONTRIBUTING.md](CONTRIBUTING.md). El proyecto usa conventional commits, branch naming (`feat/`, `fix/`), y un flujo issue-first. Las contribuciones son bienvenidas.

---

## License

MIT © [Alejandro Arroyave](https://github.com/alebak)
