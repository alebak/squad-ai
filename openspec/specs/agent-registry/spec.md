# Agent Registry Specification

## Purpose

Defines the agent registry model, file format, remote fetch, and local cache. The registry is the source of truth for available AI coding agents.

## Requirements

### Requirement: Agent Type

The Agent, InstallCmd, RuntimeDep, Checksum types, and the ConfigPaths field SHALL round-trip through JSON. Fields: `id`, `name`, `description`, `version`, `detect_command`, `install` (method, url, command), `config_paths` (optional string array), `dependencies` (runtime, optional min_version), `checksum` (sha256, verified_at), `tags`, `added_at`. Optional fields (`checksum`, `config_paths`, `min_version`) SHALL use `omitempty`.
(Previously: no `config_paths` field existed)

#### Scenario: Round-trip serialization

- GIVEN an Agent with all fields populated
- WHEN marshaled to JSON and unmarshaled back
- THEN all fields match including nested InstallCmd, RuntimeDep, Checksum, and ConfigPaths

#### Scenario: Optional fields omitted

- GIVEN an Agent with nil Checksum, nil ConfigPaths, and Dependencies containing only runtime without min_version
- WHEN marshaled to JSON
- THEN `checksum`, `config_paths`, and `min_version` fields are absent from output

### Requirement: Registry File

`registry/agents.json` SHALL be valid JSON, parse to `[]Agent`, and contain exactly 7 MVP agents with correct IDs.

#### Scenario: Parse all 7 agents

- GIVEN the `registry/agents.json` file
- WHEN parsed into `[]Agent`
- THEN it yields 7 agents with IDs: `claude-code`, `opencode`, `pi`, `codex`, `antigravity-cli`, `gemini-cli`, `gentle-ai`
- AND each has non-empty ID, Name, Install.Command, and Description

#### Scenario: Empty agent list

- GIVEN a JSON file containing `[]`
- WHEN parsed into `[]Agent`
- THEN it yields zero agents with no error

### Requirement: Remote Fetch

The system SHALL fetch the registry via HTTP GET from the GitHub raw URL. It SHALL respect context cancellation. Non-200 status SHALL return an error. Network errors SHALL be wrapped with context.

#### Scenario: Successful fetch

- GIVEN the registry URL returns HTTP 200 with valid JSON
- WHEN Fetch(ctx) is called
- THEN it returns parsed `[]Agent` with no error

#### Scenario: Context cancellation

- GIVEN a cancelled context
- WHEN Fetch(ctx) is called
- THEN it returns an error wrapping the cancellation cause

#### Scenario: Non-200 response

- GIVEN the registry URL returns HTTP 404
- WHEN Fetch(ctx) is called
- THEN it returns an error indicating non-200 status

#### Scenario: Network error

- GIVEN the registry URL is unreachable
- WHEN Fetch(ctx) is called
- THEN the error message wraps the underlying network failure

### Requirement: Local Cache

SaveCache SHALL write valid JSON to disk. LoadCache SHALL read it back producing the same Agent list. IsStale SHALL return true when file is older than maxAge, false otherwise, and false when file does not exist.

#### Scenario: Cache round-trip

- GIVEN an in-memory list of 7 agents
- WHEN SaveCache writes to a temp path and LoadCache reads it back
- THEN the loaded list is identical to the original

#### Scenario: Stale cache

- GIVEN a cache file with mtime 25 hours ago
- WHEN IsStale(cachePath, 24*time.Hour) is called
- THEN it returns true

#### Scenario: Fresh cache

- GIVEN a cache file with mtime 1 hour ago
- WHEN IsStale(cachePath, 24*time.Hour) is called
- THEN it returns false

#### Scenario: Missing cache file

- GIVEN no cache file exists at the expected path
- WHEN IsStale is called
- THEN it returns false

### Requirement: Error Handling

The system SHALL handle corrupted or unexpected registry data without panicking.

#### Scenario: Malformed JSON

- GIVEN a cache file with invalid JSON
- WHEN LoadCache is called
- THEN it returns an error wrapping the JSON parse failure

#### Scenario: Agent with empty ID

- GIVEN a JSON entry with empty `id` field
- WHEN parsed into `[]Agent`
- THEN the parser does not panic; the entry is parsed as-is (validation is caller concern)
