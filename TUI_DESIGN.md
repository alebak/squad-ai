# TUI Redesign — Uninstall Wizard & Apply Flow

## Color Palette: Catppuccin Mocha

| Role | Hex | ANSI |
|------|-----|------|
| Background | #1e1e2e | — |
| Text | #cdd6f4 | white |
| Title accent | #cba6f7 (mauve) | magenta |
| Checked | #a6e3a1 (green) | green |
| Cursor | #89b4fa (blue) | blue |
| Subdued | #6c7086 (overlay) | bright black |
| Warning | #f9e2af (yellow) | yellow |
| Error | #f38ba8 (red) | red |
| Separator | #45475a (surface1) | — |

## Issue 1: Rename "Done" → "Apply" + keyboard shortcut

```
╭──────────────────────────────────────────────────────────╮
│                                                          │
│  Squad AI (version 0.15.0)                                │
│                                                          │
│  Select Your AI Coding Agents                            │
│                                                          │
│  ○ select all                                            │
│                                                          │
│  ◉ Claude Code                                           │
│  ◉ OpenCode                                              │
│  ○ Pi (Node.js is required)                              │
│  ○ Codex CLI (Node.js is required)                       │
│  ◉ Antigravity CLI                                       │
│  ◉ Gemini CLI                                            │
│  ◉ Gentle AI                                             │
│                                                          │
│  ─────────────────────────────────────────────────────── │
│  ▸ Apply                                                 │
│                                                          │
│  ↑↓/jk navigate • space/enter toggle • a apply • q quit  │
│                                                          │
╰──────────────────────────────────────────────────────────╯
```

- "Done" renamed to "Apply"
- Separator line (───) above Apply
- Press `a` to apply from anywhere
- Enter on Apply also applies

## Issue 2: No changes → dialog

```
╭──────────────────────────────────────────────────────────╮
│                                                          │
│  Squad AI (version 0.15.0)                                │
│                                                          │
│  Select Your AI Coding Agents                            │
│                                                          │
│  ○ select all                                            │
│                                                          │
│  ... agents ...                                          │
│                                                          │
│  ─────────────────────────────────────────────────────── │
│  ▸ Apply                                                 │
│                                                          │
│  ╭─────────────────────────────────────────────────────╮ │
│  │                                                     │ │
│  │  No changes to apply.                               │ │
│  │                                                     │ │
│  │  Press enter to continue...                         │ │
│  │                                                     │ │
│  ╰─────────────────────────────────────────────────────╯ │
│                                                          │
│  ↑↓/jk navigate • space/enter toggle • a apply • q quit  │
│                                                          │
╰──────────────────────────────────────────────────────────╯
```

- Overlay dialog inside the TUI box
- "No changes to apply. Press enter to continue..."
- Press enter → returns to agent list

## Issue 3: Uninstall wizard (inline)

When user presses Apply and installed agents were deselected, show
a step-by-step wizard INSIDE the TUI box:

```
╭──────────────────────────────────────────────────────────╮
│                                                          │
│  Squad AI (version 0.15.0)                                │
│                                                          │
│    Step 1 of 2 — Claude Code                             │
│                                                          │
│    This agent is currently installed.                    │
│    Choose an action:                                     │
│                                                          │
│    ▸ Uninstall app only                                  │
│      Uninstall app + config data                         │
│      Keep installed (skip)                               │
│                                                          │
│                                                          │
│     enter select • ↑↓ navigate • n next • b back         │
│                                                          │
│                                                          │
╰──────────────────────────────────────────────────────────╯
```

```
Step 1 of 2 — Claude Code
  ▸ Uninstall app only
    Uninstall app + config data
    Keep installed (skip)

Step 2 of 2 — OpenCode
    Uninstall app only
  ▸ Uninstall app + config data
    Keep installed (skip)

After all steps, show Apply at bottom again:

  ───────────────────────────────────────────────────────
  ▸ Apply
```

### Navigation inside wizard:
- `↑↓` / `jk` — highlight option
- `enter` — select highlighted option
- `n` — next step
- `b` — previous step
- `q` — cancel wizard, return to agent list

### After wizard completes:
- Show "Apply" item at bottom again
- User must press Apply/Enter to confirm everything
- If they press `q` instead, discard wizard changes and return to agent list

## Full flow

```
squad
  │
  ▼
Agent list (with select all + Apply)
  │
  ├─ Apply (no changes) → "No changes" dialog → back to list
  │
  ├─ Apply (new agents selected) → install → progress → updated list
  │
  └─ Apply (installed deselected) → Uninstall Wizard (step by step) → Apply → updated list
```
