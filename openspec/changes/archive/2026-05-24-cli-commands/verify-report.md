# Verify Report: CLI Commands (install + list)

## Build

```
$ go build ./cmd/squad
→ SUCCESS (no output, binary created)
```

## Lint

```
$ go vet ./...
→ SUCCESS (no issues)
```

## Tests

```
$ go test ./... -v -count=1
→ 47 tests PASS (all packages)
```

### CLI package tests (new + existing): 17 PASS
| Test | Status |
|------|--------|
| TestInstallCommand_DefaultWithConfig | PASS |
| TestInstallCommand_DefaultWithConfig_AlreadyInstalled | PASS |
| TestInstallCommand_AgentsFlag | PASS |
| TestInstallCommand_AllFlag | PASS |
| TestInstallCommand_MixedFlagsError | PASS |
| TestInstallCommand_NoConfig | PASS |
| TestInstallCommand_RegistryFetchFailure | PASS |
| TestInstallCommand_PartialFailure | PASS |
| TestInstallCommand_EmptyRegistry | PASS |
| TestParseAgentIDs (5 sub-tests) | PASS |
| TestFilterAgentsByID (5 sub-tests) | PASS |
| TestListCommand_NormalOutput | PASS |
| TestListCommand_EmptyRegistry | PASS |
| TestListCommand_RegistryFetchFailure | PASS |
| TestListCommand_AllAvailable | PASS |
| TestVersionCommand_Output (2 sub-tests) | PASS |

### Other packages (existing tests, all pass):
- config: 11 PASS
- installer: 16 PASS
- registry: 11 PASS
- runtime: 14 PASS (including real Go/Python detection)

### Scenarios verified

| Scenario | Status |
|----------|--------|
| S1: Default install with config | PASS |
| S2: `--agents` flag | PASS |
| S3: `--all` flag | PASS |
| S4: No config → no-op | PASS |
| S5: Mixed flags error | PASS |
| S6: Registry fetch failure | PASS |
| S1 (list): Normal table output | PASS |
| S2 (list): Empty registry | PASS |
| S3 (list): Registry fetch failure | PASS |

## Coverage Notes

- All handler paths covered via injected mock functions
- Edge cases: empty registry, partial failure, already installed, blocked by runtime, mutually exclusive flags
- Real runtime detection (go, python3) tested and passing in the existing runtime package

## Conclusion

All requirements from specs are implemented and verified. No regressions. Ready for archive.
