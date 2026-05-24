# User Config Specification

## Purpose

Defines the user configuration model, file persistence, and default values. The config stores the user's agent selection and installation preferences.

## Requirements

### Requirement: Config Type

The Config and InstallOptions types SHALL round-trip through JSON. Fields: `selected_agents` ([]string), `registry_last_checked` (RFC3339 string), `registry_known_agents` ([]string), `install_options` (silent bool, prefer_pnpm bool).

#### Scenario: Round-trip serialization

- GIVEN a Config with two selected agents, a past timestamp, and Silent=true
- WHEN marshaled to JSON and unmarshaled back
- THEN all fields match the original values

#### Scenario: Zero-value defaults on unmarshal

- GIVEN the JSON `{}`
- WHEN unmarshaled into Config
- THEN SelectedAgents and RegistryKnown are nil, RegistryLastCheck is empty, Silent and PreferPnpm are false

### Requirement: Config Path

ConfigPath SHALL return `$HOME/.config/squad-ai/config.json`. It SHALL create parent directories with 0755 permissions when they do not exist.

#### Scenario: Default path

- GIVEN HOME is /home/user
- WHEN ConfigPath() is called
- THEN it returns `/home/user/.config/squad-ai/config.json`

#### Scenario: Directory creation

- GIVEN `~/.config/squad-ai/` does not exist
- WHEN ConfigPath() is called
- THEN the directory is created with 0755 permissions

### Requirement: Load

Load SHALL read `config.json` from disk and unmarshal it. If the file does not exist, it SHALL return DefaultConfig with no error. Malformed JSON SHALL return an error.

#### Scenario: Existing config

- GIVEN a config.json with valid content
- WHEN Load() is called
- THEN it returns a Config with matching field values

#### Scenario: Missing config file

- GIVEN no config.json exists
- WHEN Load() is called
- THEN it returns DefaultConfig with no error

#### Scenario: Malformed JSON

- GIVEN config.json contains invalid JSON
- WHEN Load() is called
- THEN it returns an error wrapping the JSON parse failure

### Requirement: Save

Save SHALL write config.json atomically using temp file + os.Rename. File permissions SHALL be 0644. Parent directories SHALL be created if missing.

#### Scenario: Atomic write

- GIVEN a Config value
- WHEN Save() is called
- THEN the file is written atomically with 0644 permissions
- AND no partial write is visible to concurrent readers

#### Scenario: Missing parent directory

- GIVEN `~/.config/squad-ai/` does not exist
- WHEN Save() is called
- THEN the directory is created and config.json is written successfully

### Requirement: DefaultConfig

DefaultConfig SHALL return sensible defaults: empty selected_agents, zero timestamp, empty known_agents, Silent=true, PreferPnpm=true.

#### Scenario: Default values

- WHEN DefaultConfig() is called
- THEN SelectedAgents is empty, RegistryLastCheck is zero-value, RegistryKnown is empty
- AND InstallOptions.Silent is true, InstallOptions.PreferPnpm is true

### Requirement: Edge Cases

The system SHALL handle config path and permission errors gracefully.

#### Scenario: Read-only directory

- GIVEN the config directory is read-only
- WHEN Save() is called
- THEN it returns an error wrapping the write failure
- AND no partial file remains

#### Scenario: Concurrent writes

- GIVEN two goroutines call Save() concurrently
- WHEN both complete
- THEN the last writer wins; both writes produce valid JSON with no corruption

#### Scenario: Missing HOME

- GIVEN HOME environment variable is unset
- WHEN ConfigPath() is called
- THEN it returns an error indicating HOME is not set
