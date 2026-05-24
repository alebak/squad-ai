# Spec: Agent Uninstall (Delta)

Delta spec extending `openspec/specs/remove/spec.md` with uninstall support.

## References

- Issue #19 — Agent uninstall support
- PRD §8 — `squad remove` table entry
- `openspec/specs/remove/spec.md`

## Requirements

### Requirement: Uninstall Field in Registry

The `InstallCmd` struct SHALL include an optional `UninstallCmd` string field mapped to JSON key `"uninstall"`.

#### Scenario: Default (empty) uninstall field

- GIVEN an `InstallCmd` without an `"uninstall"` key
- WHEN the JSON is unmarshalled
- THEN `UninstallCmd` SHALL be empty string
- AND existing install behavior SHALL be unaffected

### Requirement: UninstallAgent Function

The installer package SHALL export `UninstallAgent(agent registry.Agent) error`.

Resolution order:
1. If `agent.Install.UninstallCmd` is non-empty → execute it via `sh -c`
2. Else if method is `npm_install` → derive `npm uninstall -g <package>` from install command
3. Else if method is `curl_bash` → use `exec.LookPath` + `os.Remove` to delete the binary
4. Else → return error: no uninstall method defined

#### Scenario: Explicit uninstall command

- GIVEN an agent with `UninstallCmd: "npm uninstall -g @openai/codex"`
- WHEN `UninstallAgent` is called
- THEN the command SHALL execute via `sh -c`
- AND the command SHALL be validated (no null bytes)

#### Scenario: npm_install method fallback

- GIVEN an agent with `Method: "npm_install"` and `Command: "npm i -g @openai/codex"`
- AND `UninstallCmd` is empty
- WHEN `UninstallAgent` is called
- THEN the package name `@openai/codex` SHALL be extracted from the install command
- AND `npm uninstall -g @openai/codex` SHALL be executed

#### Scenario: curl_bash method fallback

- GIVEN an agent with `Method: "curl_bash"` and `DetectCmd: "opencode"`
- AND `UninstallCmd` is empty
- WHEN `UninstallAgent` is called
- THEN `exec.LookPath("opencode")` SHALL resolve the binary path
- AND `os.Remove(path)` SHALL delete the binary

#### Scenario: custom method with no uninstall

- GIVEN an agent with `Method: "custom"` and no `UninstallCmd`
- WHEN `UninstallAgent` is called
- THEN an error SHALL be returned indicating no uninstall method is defined

#### Scenario: curl_bash with empty detect_cmd

- GIVEN an agent with `Method: "curl_bash"` and empty `DetectCmd`
- WHEN `UninstallAgent` is called
- THEN an error SHALL be returned indicating `detect_command` is required

#### Scenario: Null byte in uninstall command

- GIVEN an `UninstallCmd` containing a null byte
- WHEN `UninstallAgent` is called
- THEN an error SHALL be returned indicating the command is invalid

### Requirement: `squad remove --uninstall` Flag

`squad remove <id> --uninstall` SHALL uninstall the agent binary before removing it from the config.

#### Scenario: Uninstall an installed agent

- GIVEN an agent installed on the system and in `selected_agents`
- WHEN the user runs `squad remove <id> --uninstall --force`
- THEN the agent SHALL be uninstalled (binary removed)
- AND the agent SHALL be removed from `selected_agents`
- AND output SHALL contain an uninstall confirmation message

#### Scenario: Uninstall when agent not on system

- GIVEN an agent in `selected_agents` but NOT installed (binary missing)
- WHEN the user runs `squad remove <id> --uninstall --force`
- THEN the command SHALL print a notice that the agent is not installed
- AND the agent SHALL still be removed from `selected_agents`

#### Scenario: Uninstall without --uninstall flag (backward compat)

- GIVEN an agent installed on the system and in `selected_agents`
- WHEN the user runs `squad remove <id>` (no `--uninstall`)
- THEN the agent SHALL NOT be uninstalled
- AND the agent SHALL be removed from `selected_agents`
- AND output SHALL contain the note about the agent still being installed

#### Scenario: Confirmation prompt

- GIVEN the user runs `squad remove <id> --uninstall` WITHOUT `--force`
- WHEN the agent is installed
- THEN the command SHALL display the uninstall action and prompt for confirmation
- AND SHALL only proceed on affirmative response

#### Scenario: Agent not in registry

- GIVEN the user runs `squad remove <id> --uninstall --force`
- AND `<id>` is not in the remote registry
- WHEN the registry is fetched
- THEN the command SHALL print a warning about the agent not being found in the registry
- AND the agent SHALL still be removed from `selected_agents`

### Requirement: Registry Data Updates

The `registry/agents.json` SHALL include explicit `"uninstall"` commands for all npm_install agents.

#### Scenario: Agent with explicit uninstall

- GIVEN an agent entry with `"install": {"method": "npm_install", ...}`
- WHEN the agent has an `"uninstall"` field
- THEN `UninstallAgent` SHALL prefer the explicit command over the fallback

Expected entries:
- `pi` → `"uninstall": "npm uninstall -g @earendil-works/pi-coding-agent"`
- `codex` → `"uninstall": "npm uninstall -g @openai/codex"`
- `gemini-cli` → `"uninstall": "npm uninstall -g @google/gemini-cli"`
- `curl_bash` agents → no `"uninstall"` key (use fallback)
