# Squad AI

**Gestioná tu escuadrón de coding agents dentro de dev containers.**

[![Go 1.24+](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey)](https://github.com/alebak/squad-ai)

> Read this document in [English](README.md).

---

## Qué hace

Squad AI es un CLI en Go que instala y gestiona coding agents (Claude Code, OpenCode, Pi, Codex y más) dentro de dev containers. Reemplaza los scripts `post-create.sh` manuales por una herramienta unificada.

**Antes**: copiar snippets de `curl | bash` entre proyectos, mantenerlos sincronizados a mano.  
**Después**: `squad add` → elegís tus agentes → listo. Un solo comando en `postCreateCommand`.

## Instalación rápida

```bash
curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
```

Esto descarga el binario a `~/.local/bin/squad`. Agregalo a tu PATH si hace falta:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

## Uso

```bash
# Primera vez: elegí tus agentes interactivamente
squad add

# Instalá los agentes seleccionados en silencio
squad install

# Instalá agentes específicos (sin TUI)
squad install --agents claude-code,opencode,gentle-ai

# Mirá qué está instalado y qué hay disponible
squad list

# Actualizá el registro de agentes
squad update
```

## Integración con dev containers

Agregá a tu `.devcontainer/devcontainer.json`:

```json
{
  "postCreateCommand": "squad install"
}
```

O usá un `post-create.sh` para la primera vez:

```bash
#!/bin/bash
set -e

if ! command -v squad &>/dev/null; then
  curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/scripts/install.sh | bash
  export PATH="$HOME/.local/bin:$PATH"
fi

squad install
```

La configuración persiste entre reconstrucciones del container si `$HOME` se monta como volumen.

## Agentes soportados

| Agente | Instalación | Runtime |
|--------|------------|---------|
| Claude Code | curl \| bash | ninguno |
| OpenCode | curl \| bash | ninguno |
| Pi | curl \| bash | ninguno |
| Codex CLI | npm | Node.js |
| Antigravity CLI | curl \| bash | ninguno |
| Gemini CLI | npm | Node.js |
| Gentle AI | curl \| bash | ninguno |

Se agregan agentes nuevos vía PR a `registry/agents.json`. Squad AI te avisa cuando hay agentes nuevos disponibles.

## Cómo funciona

1. **Registro** — `registry/agents.json` en este repo lista los agentes disponibles con comandos de instalación, dependencias de runtime y checksums SHA-256.
2. **Cache local** — squad descarga y cachea el registro en `~/.config/squad-ai/registry.cache.json`.
3. **Detección** — con `squad list` verifica qué agentes ya están instalados y qué runtimes faltan.
4. **Instalación** — `squad install` ejecuta los comandos, captura logs, verifica checksums y guarda tu selección.
5. **TUI** — `squad add` abre una interfaz de terminal donde explorás, seleccionás e instalás agentes.

## Comandos

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

## Contribuir

Ver [CONTRIBUTING.md](CONTRIBUTING.md). El proyecto usa conventional commits, branch naming (`feat/`, `fix/`), y un flujo issue-first. Las contribuciones son bienvenidas.

¿Preguntas? Usá [Discussions](https://github.com/alebak/squad-ai/discussions), no los issues.

## Licencia

MIT © [Alebak](https://github.com/alebak)
