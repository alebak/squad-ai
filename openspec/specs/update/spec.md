# Spec: `squad update` Command

## References

- PRD §8 (Comandos del CLI) — `squad update` table entry
- PRD §7.4 (Modo interactivo manual) — update usage context

## Requirements

### Requirement: Force Registry Update

`squad update` SHALL force a remote registry re-fetch, save the result to the local cache, and print a confirmation message.

#### Scenario: Successful update

- GIVEN the registry is reachable
- WHEN the user runs `squad update`
- THEN the registry is fetched from remote
- AND the result is saved to the local cache
- AND the output contains a confirmation message with agent count

### Requirement: Registry Fetch Failure

If the remote registry cannot be fetched, `squad update` SHALL error with a message explaining the failure.

#### Scenario: Update fails

- GIVEN the registry is unreachable
- WHEN the user runs `squad update`
- THEN the command errors explaining the network failure
- AND the existing cache is NOT modified
