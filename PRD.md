# Squad AI — Product Requirements Document

**Version:** 0.1.0
**Author:** Alejandro Arroyave
**Date:** 2026-05-23
**Repository:** https://github.com/alebak/squad-ai
**License:** MIT

---

## 1. Resumen ejecutivo

Squad AI es un CLI en Go que gestiona la instalación de coding agents y herramientas de desarrollo dentro de containers (y potencialmente cualquier entorno Linux). Se distribuye como binario estático, se invoca desde `postCreateCommand` en devcontainers, y resuelve el problema de mantener scripts `post-create.sh` duplicados y divergentes entre proyectos.

**Analogía:** Un gestor de paquetes especializado en coding agents — "brew para agentes de IA dentro de containers".

---

## 2. Problema

En flujos de desarrollo con Dev Containers, cada proyecto requiere un script `post-create.sh` que instala manualmente los coding agents deseados (Claude Code, OpenCode, gentle-ai, etc.). Este patrón presenta los siguientes problemas:

- **Duplicación**: el mismo bloque `if command -v ... ; then ... ; fi` se copia entre proyectos.
- **Divergencia**: al agregar o remover un agente, hay que editar N scripts en N repositorios.
- **Rigidez**: no hay forma de personalizar la selección por desarrollador sin editar el script.
- **Descubrimiento nulo**: cuando aparece un agente nuevo (Pi, Codex, Antigravity CLI), no hay mecanismo de notificación ni catálogo centralizado.
- **Fragilidad**: los scripts `curl | bash` de terceros pueden cambiar sin aviso, rompiendo instalaciones o introduciendo vulnerabilidades (supply chain attacks).

---

## 3. Solución

Un CLI llamado `squad` que:

1. Mantiene un **registry centralizado** de agentes disponibles con sus métodos de instalación, dependencias de runtime, y checksums.
2. Permite al desarrollador **seleccionar su tropa** de agentes mediante un TUI interactivo.
3. Se integra en el **ciclo de vida del devcontainer** (`postCreateCommand`) para instalar automáticamente los agentes seleccionados.
4. **Detecta agentes nuevos** en el registry y notifica al desarrollador.
5. **Verifica dependencias de runtime** antes de instalar (Node.js, Go, Python, etc.) y bloquea agentes cuyo runtime no está presente.
6. **Instala silenciosamente** cada agente, capturando stdout/stderr y mostrando su propia barra de progreso.

---

## 4. Usuarios objetivo

- Desarrolladores que trabajan con Dev Containers en VS Code.
- Equipos que necesitan estandarizar qué herramientas de IA están disponibles en sus entornos de desarrollo.
- Desarrolladores individuales que quieren un catálogo actualizado de coding agents sin mantener scripts manuales.

---

## 5. Stack técnico

| Componente | Tecnología |
|---|---|
| Lenguaje | Go |
| Estructura de comandos | Cobra |
| TUI interactivo | Bubbletea (Charm) |
| Styling del TUI | Lip Gloss (Charm) |
| Componentes TUI | Bubbles (spinners, progress bars, listas) |
| Build y distribución | GoReleaser |
| Testing | Go estándar (`testing`) + testify |
| Linting | golangci-lint |

---

## 6. Arquitectura de alto nivel

### 6.1. Distribución

Squad AI se distribuye como binario estático. Se instala con un one-liner:

```bash
curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/install.sh | bash
```

El binario se descarga a `~/.local/bin/squad` (o la ruta apropiada según el sistema).

### 6.2. Configuración del usuario

Toda la configuración vive en `~/.config/squad-ai/`:

```
~/.config/squad-ai/
├── config.json        # Manifiesto del usuario: agentes seleccionados
└── registry.cache.json # Cache local del registry remoto
```

Sigue la convención XDG Base Directory. No hay archivos de configuración a nivel de proyecto — Squad AI es herramienta del desarrollador, no del proyecto.

### 6.3. Registry remoto

Archivo JSON hospedado en el repositorio de Squad AI en GitHub:

```
https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json
```

Estructura de cada entrada del registry:

```json
{
  "id": "claude-code",
  "name": "Claude Code",
  "description": "Anthropic's AI coding agent",
  "version": "latest",
  "detect_command": "claude",
  "install": {
    "method": "curl_bash",
    "url": "https://claude.ai/install.sh",
    "command": "curl -fsSL https://claude.ai/install.sh | bash"
  },
  "dependencies": [
    {
      "runtime": "none"
    }
  ],
  "checksum": {
    "sha256": "abc123...",
    "verified_at": "2026-05-23T00:00:00Z"
  },
  "tags": ["ai", "coding-agent", "anthropic"],
  "added_at": "2026-05-23T00:00:00Z"
}
```

Campos clave por agente:

- **id**: identificador único (slug).
- **name**: nombre para mostrar en el TUI.
- **description**: descripción breve.
- **version**: versión conocida o `"latest"`.
- **detect_command**: comando para verificar si ya está instalado (`command -v <detect_command>`).
- **install.method**: tipo de instalación (`curl_bash`, `go_install`, `npm_install`, `binary_download`, `custom`).
- **install.command**: comando completo de instalación.
- **dependencies**: lista de runtimes requeridos. Cada dependencia especifica `runtime` (`node`, `go`, `python`, `none`) y opcionalmente `min_version`.
- **checksum**: hash SHA-256 del script de instalación para verificación de integridad.
- **tags**: etiquetas para filtrado y categorización.
- **added_at**: fecha de incorporación al registry (para detectar novedades).

### 6.4. Manifiesto del usuario (`config.json`)

```json
{
  "selected_agents": ["claude-code", "opencode", "gentle-ai"],
  "registry_last_checked": "2026-05-23T10:00:00Z",
  "registry_known_agents": ["claude-code", "opencode", "gentle-ai"],
  "install_options": {
    "silent": true,
    "prefer_pnpm": true
  }
}
```

- **selected_agents**: lista de IDs de agentes que el usuario quiere tener instalados.
- **registry_last_checked**: timestamp de la última consulta al registry.
- **registry_known_agents**: lista de agentes que el usuario ya conoce (para detectar novedades).
- **install_options**: preferencias globales de instalación.

---

## 7. Modos de operación

### 7.1. Primera ejecución (sin manifiesto)

Contexto: el desarrollador acaba de instalar Squad AI y lo ejecuta por primera vez (ya sea manualmente o desde `postCreateCommand`).

Flujo:

1. Squad detecta que no existe `~/.config/squad-ai/config.json`.
2. Descarga el registry remoto.
3. Lanza el TUI con la lista de agentes disponibles.
4. Todos los agentes compatibles aparecen seleccionados por defecto (checkbox marcado).
5. Agentes cuyo runtime no está presente aparecen deshabilitados con indicación del runtime faltante (ejemplo: `⛔ Claude Code  — requiere Node.js 18+`). *Nota: Claude Code actualmente no requiere Node.js, esto es un ejemplo ilustrativo.*
6. El usuario desmarca los que no quiere y confirma.
7. Squad instala los seleccionados secuencialmente, mostrando barra de progreso por cada uno.
8. Genera `config.json` con la selección.

### 7.2. Ejecuciones posteriores (manifiesto existe)

Contexto: el container se reconstruye, `postCreateCommand` ejecuta `squad` de nuevo.

Flujo:

1. Squad lee `config.json`.
2. Para cada agente seleccionado, verifica si ya está instalado (`command -v`).
3. Instala silenciosamente los que faltan, con barra de progreso.
4. Consulta el registry remoto (si no lo ha hecho en las últimas 24h, configurable).
5. Compara `registry_known_agents` con los agentes actuales del registry.
6. Si hay agentes nuevos, muestra un mensaje informativo al final:

```
✅ Claude Code instalado
✅ OpenCode instalado
✅ gentle-ai instalado

ℹ️  Nuevos agentes disponibles en el registry:
   • Antigravity CLI — AI-powered coding agent by Google
   • Pi Terminal — Personal AI assistant
   Ejecuta 'squad add' para explorarlos.
```

7. No lanza TUI. El flujo es completamente no-interactivo.

### 7.3. Override por parámetros

Para personalización dentro de `post-create.sh` sin TUI:

```bash
# Instalar agentes específicos directamente (sin TUI)
squad install --agents claude-code,opencode,gentle-ai

# Instalar todos los del registry que sean compatibles
squad install --all

# Instalar solo los del manifiesto existente (por defecto)
squad install
```

Cuando se usa `--agents`, Squad:

1. Salta el TUI.
2. Instala los agentes listados (verificando dependencias).
3. Actualiza el manifiesto con la selección.

### 7.4. Modo interactivo manual

El usuario ejecuta `squad` manualmente (sin estar en `postCreateCommand`):

```bash
# Abrir TUI para gestionar agentes
squad

# Agregar agentes (abre TUI con los no instalados)
squad add

# Listar agentes instalados y disponibles
squad list

# Ver detalle de un agente
squad info claude-code

# Actualizar el registry local
squad update
```

---

## 8. Comandos del CLI

| Comando | Descripción |
|---|---|
| `squad` | Sin argumentos: si no hay manifiesto, lanza TUI de primera vez. Si hay manifiesto, ejecuta instalación silenciosa (modo postCreate). |
| `squad install` | Instala agentes del manifiesto. Con `--agents <ids>` instala específicos. Con `--all` instala todos los compatibles. |
| `squad add` | Abre TUI mostrando agentes disponibles no instalados. |
| `squad remove <id>` | Remueve un agente del manifiesto (no lo desinstala del sistema). |
| `squad list` | Lista agentes: instalados, seleccionados, disponibles, bloqueados. |
| `squad info <id>` | Muestra detalle de un agente (versión, runtime, método de instalación). |
| `squad update` | Fuerza actualización del registry desde GitHub. |
| `squad version` | Muestra la versión de Squad AI. |

---

## 9. Verificación de dependencias de runtime

Antes de instalar un agente, Squad verifica sus dependencias:

| Runtime | Detección | Versión |
|---|---|---|
| Node.js | `node --version` | Parseo de semver, comparación con `min_version` |
| Go | `go version` | Parseo de `go1.x.y` |
| Python | `python3 --version` | Parseo de semver |
| Rust/Cargo | `cargo --version` | Parseo de semver |
| None | — | Siempre compatible (binario estático) |

Si un runtime falta:

- En modo TUI: el agente aparece con ícono de bloqueo y mensaje del runtime requerido. No se puede seleccionar.
- En modo `--agents`: Squad muestra error y salta ese agente sin detener el resto.
- Squad **no instala runtimes**. Solo informa. La instalación del runtime es responsabilidad del Dockerfile o las features del devcontainer.

---

## 10. Instalación silenciosa con progreso

Cada agente se instala con stdout y stderr redirigidos. Squad muestra su propia interfaz de progreso:

```
📦 Instalando Claude Code...
  ████████████████████░░░░░░░░░░  67%

✅ Claude Code instalado (12s)
📦 Instalando OpenCode...
  ██████████░░░░░░░░░░░░░░░░░░░░  33%
```

El progreso es **cosmético** (avance gradual animado que salta a 100% al completar), ya que los scripts de instalación de terceros no reportan progreso de forma parseable. La barra avanza a un ritmo estimado basado en tiempos históricos promedio por agente, almacenados en el registry.

Si la instalación falla:

```
❌ OpenCode — instalación fallida
   Detalle: exit code 1
   Log: ~/.config/squad-ai/logs/opencode-2026-05-23.log
```

Los logs completos de cada instalación se guardan en `~/.config/squad-ai/logs/` para diagnóstico.

---

## 11. Seguridad e integridad

### 11.1. MVP

- El registry incluye el campo `checksum.sha256` por cada script de instalación.
- Squad descarga el script, calcula su hash, y lo compara con el valor del registry antes de ejecutarlo.
- Si el hash no coincide, Squad rechaza la instalación y muestra advertencia.

### 11.2. Post-MVP (futuro)

- Lockfile (`~/.config/squad-ai/squad.lock`) con hashes verificados de cada script.
- Detección de cambios en scripts entre ejecuciones (alerta si un script cambió desde la última verificación).
- Firma GPG del registry.
- Opción de hospedar copias verificadas de los scripts de instalación en el repo de Squad AI.

---

## 12. Integración con devcontainers

### 12.1. Uso típico en `post-create.sh`

```bash
#!/bin/bash
set -e

# Instalar Squad AI si no está presente
if ! command -v squad &>/dev/null; then
  curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/install.sh | bash
  export PATH="$HOME/.local/bin:$PATH"
fi

# Ejecutar Squad (primera vez: TUI, siguientes: silencioso)
squad
```

### 12.2. Uso con selección predefinida (sin TUI)

```bash
#!/bin/bash
set -e

if ! command -v squad &>/dev/null; then
  curl -fsSL https://raw.githubusercontent.com/alebak/squad-ai/main/install.sh | bash
  export PATH="$HOME/.local/bin:$PATH"
fi

# Instalar agentes específicos sin interacción
squad install --agents claude-code,opencode,gentle-ai
```

### 12.3. Referencia en `devcontainer.json`

```json
{
  "postCreateCommand": "bash ./.devcontainer/post-create.sh"
}
```

### 12.4. Persistencia del home

La configuración en `~/.config/squad-ai/` persiste entre rebuilds del container siempre que el home del usuario se monte como volumen (práctica estándar en devcontainers). Esto significa que el TUI de selección solo aparece una vez, no en cada rebuild.

---

## 13. Registry inicial (MVP)

Agentes incluidos en el registry del MVP:

| ID | Nombre | Método de instalación | Runtime |
|---|---|---|---|
| `claude-code` | Claude Code | `curl \| bash` | Ninguno (binario) |
| `opencode` | OpenCode | `curl \| bash` | Ninguno (binario) |
| `gentle-ai` | gentle-ai | `curl \| bash` | Ninguno (binario) |
| `codex` | Codex CLI | `npm install -g` | Node.js 22+ |
| `gemini-cli` | Gemini CLI | `npm install -g` | Node.js 18+ |
| `aider` | Aider | `pip install` | Python 3.10+ |

*Esta lista se ampliará a medida que aparezcan nuevos agentes. Contribuciones vía PR al registry son bienvenidas.*

---

## 14. Fases de desarrollo

### Fase 1 — MVP

- Estructura del proyecto con Cobra.
- Registry remoto en JSON (hospedado en el repo).
- Lectura y cache del registry.
- Detección de agentes instalados (`command -v`).
- Verificación de dependencias de runtime.
- TUI de selección (primera ejecución) con Bubbletea.
- Instalación silenciosa con barra de progreso cosmética.
- Persistencia del manifiesto en `~/.config/squad-ai/config.json`.
- Modo no-interactivo con `--agents`.
- Notificación de agentes nuevos.
- Logging de instalaciones a archivo.
- Comando `squad list`.
- GoReleaser para builds multiplataforma (linux/amd64, linux/arm64).
- Script `install.sh` para distribución.

### Fase 2 — Seguridad y robustez

- Verificación de checksums SHA-256.
- Lockfile de integridad.
- Detección de cambios en scripts de instalación.
- Comandos `squad info`, `squad update`.
- Soporte para método de instalación `binary_download` (descarga directa de binarios sin `curl | bash`).
- Tiempos estimados de instalación por agente para progreso más realista.

### Fase 3 — Comunidad y extensibilidad

- Contribución de agentes vía PR (plantilla y validación).
- Categorización y filtrado por tags en el TUI.
- Soporte para registries alternativos (empresariales/privados).
- Comando `squad doctor` para diagnóstico del entorno.
- Auto-actualización del binario de Squad AI.

---

## 15. Decisiones de diseño

| Decisión | Razonamiento |
|---|---|
| Configuración en `~/.config/` (no en el proyecto) | Squad es herramienta del desarrollador, no del proyecto. Cada quien elige su tropa. |
| Go como lenguaje | Binario estático sin dependencias de runtime. Mismo ecosistema que lazyvol, jkit, y gentle-ai. |
| Registry en el repo de GitHub | Simplicidad: un PR para agregar un agente. Sin infraestructura de backend. |
| Progreso cosmético | Los scripts de instalación de terceros no reportan progreso. Una barra animada es mejor UX que un spinner estático. |
| No instalar runtimes | Reducir scope. Los runtimes son decisión del Dockerfile o features del devcontainer, no del gestor de agentes. |
| Solo informar novedades (no instalar automáticamente) | Respetar la agencia del desarrollador. Notificar sin imponer. |

---

## 16. Fuera de alcance (explícitamente)

- Instalación de runtimes (Node.js, Go, Python).
- Gestión de configuración de los agentes (API keys, settings).
- Ejecución o proxy de los agentes.
- Soporte para macOS/Windows nativos (foco en containers Linux).
- Interfaz web o GUI.
- Backend o API server.

---

## 17. Métricas de éxito

- El `post-create.sh` de cualquier proyecto se reduce a ≤5 líneas.
- Agregar un agente nuevo al ecosistema requiere solo un PR al registry.
- El TUI de primera ejecución completa en <30 segundos (selección + instalación de 3 agentes).
- Zero instalaciones rotas por cambios en scripts de terceros (verificación de checksum).