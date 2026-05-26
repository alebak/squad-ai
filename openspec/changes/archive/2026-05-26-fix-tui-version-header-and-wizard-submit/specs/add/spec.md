# Delta Spec: `squad add` Command — Version Header & Wizard Submit

## Base Spec

`openspec/specs/add/spec.md`

## Changed Requirements

### Requirement: Header shows build version (REPLACES "Header shows version")

**Old requirement (line 24):**
> The header SHALL display "Squad AI (version 0.15.0)" with the mauve color for "Squad AI" and subdued color for the version.

**New requirement:**
> The header SHALL display "Squad AI (version `<build-version>`)" where `<build-version>` is the value of the `version` variable from the `cli` package, set at build time via ldflags. "Squad AI" SHALL use the mauve color. The version string SHALL use the subdued/overlay color.

**Old scenario (line 123-128):**
> - GIVEN the TUI is open
> - THEN the header contains "Squad AI (version 0.15.0)"
> - AND the "Squad AI" part uses mauve color
> - AND the version part uses subdued/overlay color

**New scenario:**
> - GIVEN the TUI is open
> - THEN the header contains "Squad AI (version " followed by the build version
> - AND the "Squad AI" part uses mauve color
> - AND the version part uses subdued/overlay color
> - AND the version SHALL NOT be a hardcoded literal (SHALL reference the cli package version)

### New Requirement: Apply on wizard summary submits immediately

**New requirement:**
> When the user presses Enter on the Apply button in the wizard summary view, the TUI SHALL immediately submit (set `isSubmitted=true`, clear `wizard`, return `tea.Quit`). The user SHALL NOT be returned to the agent list to press Apply again.

**New scenarios:**

#### Scenario: Apply on wizard summary quits immediately

- GIVEN the wizard summary view is showing
- AND the cursor is on the Apply button (cursor=0)
- WHEN the user presses Enter
- THEN `wizard` is set to nil
- AND `isSubmitted` is set to true
- AND `tea.Quit` is returned
- AND the model is NOT returned to the agent list view

#### Scenario: Back on wizard summary returns to last step (unchanged behavior)

- GIVEN the wizard summary view is showing
- AND the cursor is on the Back button (cursor=1)
- WHEN the user presses Enter
- THEN the wizard returns to the last agent step
- AND `isSubmitted` remains false

#### Scenario: Wizard choices are accessible after summary submit

- GIVEN the wizard completes and Apply is pressed on summary
- THEN `wizardOut` is populated with all agent choices
- AND `RunSelection` returns the wizard choices
