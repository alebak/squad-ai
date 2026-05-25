# Delta for Agent Registry

## MODIFIED Requirements

### Requirement: Agent Type

The Agent, InstallCmd, RuntimeDep, Checksum types, and the new ConfigPaths field SHALL round-trip through JSON. Fields: `id`, `name`, `description`, `version`, `detect_command`, `install` (method, url, command), `config_paths` (optional string array), `dependencies` (runtime, optional min_version), `checksum` (sha256, verified_at), `tags`, `added_at`. Optional fields (`checksum`, `config_paths`, `min_version`) SHALL use `omitempty`.
(Previously: no `config_paths` field existed)

#### Scenario: Config_paths round-trip

- GIVEN an Agent with `ConfigPaths: ["~/.claude", "~/.config/opencode"]`
- WHEN marshaled to JSON and unmarshaled back
- THEN `ConfigPaths` matches the original values

#### Scenario: Config_paths omitted when nil

- GIVEN an Agent with nil ConfigPaths
- WHEN marshaled to JSON
- THEN `config_paths` field is absent from output

### Requirement: Registry File

`registry/agents.json` SHALL be valid JSON, parse to `[]Agent`, and contain exactly 7 MVP agents with correct IDs. New `config_paths` entries SHALL be optional (absent for agents without config data).
(Previously: no `config_paths` field existed in the file format)

#### Scenario: Config_paths present for agents that need them

- GIVEN the `registry/agents.json` file
- WHEN parsed into `[]Agent`
- THEN agents with known config directories have non-empty `ConfigPaths`
- AND agents without config data have nil/absent `ConfigPaths`
