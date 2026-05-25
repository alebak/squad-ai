# Uninstall Specification

## Purpose

The uninstall domain manages agent removal from the system. It supports two independent operations: removing the binary (`UninstallAgent`) and removing config/data directories (`UninstallConfig`). These can be used separately or combined.

## Requirements

### Requirement: UninstallAgent

The system SHALL support agent binary removal via `UninstallAgent(agent registry.Agent) error`. Resolution order:
1. If `agent.Install.UninstallCmd` is set, execute it via `sh -c`
2. If method is `npm_install`, derive `npm uninstall -g <package>` from the install command
3. If method is `curl_bash`, resolve the binary via `exec.LookPath` and delete with `os.Remove`
4. Otherwise, return an error

#### Scenario: Explicit uninstall command

- GIVEN an agent with `UninstallCmd: "/bin/true"`
- WHEN `UninstallAgent` is called
- THEN the command executes and returns nil

#### Scenario: npm fallback derivation

- GIVEN an agent with method `npm_install` and command `npm i -g @openai/codex`
- WHEN `UninstallAgent` is called
- THEN it derives `npm uninstall -g @openai/codex`

#### Scenario: curl_bash binary removal

- GIVEN an agent with method `curl_bash` and `DetectCmd: "claude"`
- WHEN `UninstallAgent` is called
- THEN the binary at the resolved PATH is deleted

### Requirement: UninstallConfig

The system SHALL support config/data directory removal via `UninstallConfig(agent registry.Agent) error`.

#### Scenario: Remove single config dir

- GIVEN an agent with `ConfigPaths: ["~/.claude"]`
- WHEN `UninstallConfig` is called
- THEN the expanded path is removed with `os.RemoveAll`

#### Scenario: Skip non-existent path

- GIVEN an agent with `ConfigPaths: ["~/.nonexistent"]`
- WHEN `UninstallConfig` is called
- THEN nil is returned (skip, no error)

#### Scenario: Empty ConfigPaths

- GIVEN an agent with nil or empty `ConfigPaths`
- WHEN `UninstallConfig` is called
- THEN nil is returned
