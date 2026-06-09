---
title: CLI reference
description: All confire commands and global flags.
---

## `confire setup`

Installs Confire integrations for detected AI agents.

```bash
confire setup
confire setup --global
confire setup --local
```

`confire setup` is interactive by default. It prompts for scope and
supported clients.

- `--global` installs into user-level agent settings.
- `--local` installs into the nearest supported project settings, such
  as the nearest Git root.

Restart your agent after running setup.

---

## `confire start`

Starts the local Confire daemon.

```bash
confire start
```

The daemon runs in the background and handles firewall evaluation,
policy checks, tool result inspection, and local event logging. Setup
starts the daemon automatically.

---

## `confire stop`

Stops the running daemon.

```bash
confire stop
```

---

## `confire status`

Shows daemon state, connected agents, account status, policy mode,
firewall status, and sync status.

```bash
confire status
```

Use this to verify that Confire is installed and running.

---

## `confire mode`

Sets the active policy mode.

```bash
confire mode observe
confire mode balanced
confire mode strict
confire mode bypass
```

| Mode | Behavior |
|---|---|
| `observe` | Records matches without interrupting the agent |
| `balanced` | Default mode for daily work |
| `strict` | Reviews or blocks more aggressively |
| `bypass` | Temporarily disables enforcement |

---

## `confire on`

Shortcut for enabling the firewall in balanced mode.

```bash
confire on
```

Equivalent to `confire mode balanced`.

---

## `confire off`

Shortcut for bypass mode.

```bash
confire off
```

Equivalent to `confire mode bypass`. Use bypass mode carefully. For a
single approved retry, prefer `confire bypass-next`.

---

## `confire bypass-next`

Creates a one-shot approval for the next reviewed tool call.

```bash
confire bypass-next
```

The next reviewed call is allowed once, then the approval is consumed
automatically. This does not disable the firewall.

---

## `confire login`

Opens a browser to authenticate with your Confire account.

```bash
confire login
```

Login enables:

- dashboard sync,
- policy sync,
- custom rules,
- firewall history,
- Dev early access features.

The local firewall works without logging in.

---

## `confire logout`

Logs out and removes local account credentials.

```bash
confire logout
```

Built-in local rules continue working after logout.

---

## `confire policy`

Manages local and synced firewall policy.

```bash
confire policy status
confire policy test <command-or-tool>
confire policy pull
```

### `confire policy status`

Shows active mode, rule counts, custom rule status, and policy cache
information.

```bash
confire policy status
```

### `confire policy test`

Simulates a PreToolUse evaluation without running anything.

```bash
confire policy test 'git push --force'
confire policy test 'mcp__github__merge_pull_request'
confire policy test 'git status'
```

### `confire policy pull`

Fetches custom rules from Confire Cloud and updates the local policy
cache. Custom rules are part of Dev early access.

```bash
confire policy pull
```

---

## `confire events`

Shows recent local firewall events.

```bash
confire events
confire events --last 20
confire events --session current
```

Events can include:

- warnings,
- reviews,
- blocks,
- tool result findings,
- unknown MCP events,
- secret-looking value detections,
- prompt-injection-like result detections.

---

## `confire config`

Reads or writes local CLI configuration.

```bash
confire config
confire config get mode
confire config set mode=strict
confire config set dashboard_sync.enabled=false
```

See [Config file](../../configuration/config-file) for all keys.

---

## `confire reset`

Removes Confire integrations from supported agent settings and stops
the daemon. The Confire binary remains installed.

```bash
confire reset
```

---

## `confire update`

Updates the Confire CLI to the latest release.

```bash
confire update
```

---

## `confire version`

Prints the installed version.

```bash
confire version
```

---

## `confire hook`

Called automatically by supported agent integrations. Not for direct
use.

```bash
confire hook   # reads tool event from stdin, writes hook response to stdout
```

---

## Global flags

| Flag | Description |
|---|---|
| `--version` | Print version and exit |
| `--help` | Show help |
| `--local` | Use local development endpoints, if available |
