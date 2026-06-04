---
title: CLI commands
description: Complete reference for the confire command-line interface.
---

## Global flags

| Flag | Description |
|------|-------------|
| `--config <path>` | Use a specific config file instead of the default |
| `--verbose, -v` | Enable verbose output |
| `--json` | Output results as JSON (where supported) |
| `--help, -h` | Show help for a command |
| `--version` | Print the CLI version |

---

## `confire start`

Start the Confire proxy.

```bash
confire start [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--port <n>` | `4747` | Port to listen on |
| `--daemon` | false | Run as a background process |
| `--policy <path>` | `~/.confire/policy.yaml` | Policy file to load |
| `--no-sync` | false | Skip remote policy sync on start |

---

## `confire stop`

Stop a running daemon.

```bash
confire stop
```

---

## `confire status`

Show whether the proxy is running, the port it's on, and active connections.

```bash
confire status
```

---

## `confire mcp`

Start an MCP stdio transport that forwards to the proxy. Used by Claude Code.

```bash
confire mcp
```

This is an internal command — you reference it in MCP server config, you don't run it directly.

---

## `confire login`

Authenticate with Confire.

```bash
confire login [flags]
```

| Flag | Description |
|------|-------------|
| `--api-key <key>` | Authenticate using an API key (no browser) |

---

## `confire logout`

Remove stored credentials.

```bash
confire logout
```

---

## `confire test`

Send a test tool call through the proxy and show what the policy engine does with it.

```bash
confire test [flags]
```

| Flag | Description |
|------|-------------|
| `--tool <name>` | Tool name to test |
| `--server <name>` | MCP server to target |
| `--args <json>` | Tool arguments as JSON |
| `--client <name>` | Simulate a specific client (e.g. `claude-code`) |

---

## `confire policy push`

Upload a local policy file to the Confire cloud.

```bash
confire policy push <file> [flags]
```

| Flag | Description |
|------|-------------|
| `--name <name>` | Named policy slot (default: `default`) |
| `--share` | Make visible to team members |

---

## `confire policy pull`

Download a policy from the Confire cloud.

```bash
confire policy pull [flags]
```

| Flag | Description |
|------|-------------|
| `--name <name>` | Named policy to pull (default: `default`) |
| `--account <slug>` | Pull from a team account |
| `--output <path>` | Write to a file instead of stdout |

---

## `confire policy versions`

List versions of a stored policy.

```bash
confire policy versions [--name <name>]
```

---

## `confire policy rollback`

Revert to a previous policy version.

```bash
confire policy rollback --version <n>
```

---

## `confire stats`

Show token usage and savings statistics.

```bash
confire stats [flags]
```

| Flag | Description |
|------|-------------|
| `--days <n>` | Show last N days (default: 7) |
| `--json` | Output as JSON |

---

## `confire redaction test`

Test a redaction pattern against sample input.

```bash
confire redaction test --pattern <regex> --input <string>
```

---

## `confire keys`

Manage API keys.

```bash
confire keys list
confire keys create [--name <label>]
confire keys revoke <key-id>
```
