---
title: Config file
description: >-
  All Confire configuration lives in ~/.confire/config.json.
  Read and write it with `confire config`.
---

Confire's local configuration is stored at
`~/.confire/config.json`. You don't need to edit this file
directly — use the `confire config` command instead.

## Read and write config

```bash
confire config                        # show all settings
confire config get notifications.style
confire config set notifications.enabled=false
```

## Available keys

### Notifications

| Key | Values | Default | Description |
|---|---|---|---|
| `notifications.enabled` | `true` / `false` | `true` | Show 🔥 save notifications |
| `notifications.style` | `brand` / `minimal` / `off` | `brand` | Notification format |
| `notifications.min_saved_tokens` | integer | `5000` | Minimum token save to show a notification |
| `notifications.big_save_tokens` | integer | `50000` | Saves above this threshold also add a note in agent context |

`brand` style shows `🔥 Confire saved ~Xk tokens on this response`.
`minimal` shows `[confire] X → Y bytes (Z%)`. `off` silences all
output, including the context note.

### Firewall

| Key | Values | Default | Description |
|---|---|---|---|
| `mode` | `observe` / `balanced` / `strict` / `bypass` | `balanced` | Policy enforcement mode |
| `firewall_enabled` | `true` / `false` | `true` | Master firewall toggle |
| `runaway_loop_threshold` | integer | `20` | Number of identical calls within 5 min that triggers a runaway-loop block. Raise if legitimate batch work triggers false positives. |
| `call_rate_threshold` | integer | `100` | Maximum total tool calls per minute before the session is blocked. Raise for high-volume automation. |
| `ask_budget_size` | integer | `5` | Warn-action calls per session that are auto-acknowledged without surfacing to the user. Set to `0` to surface every warn as a review immediately. |

Setting `mode=bypass` or `firewall_enabled=false` both disable the
firewall. Use `confire on` / `confire off` as shortcuts.

Changes to `runaway_loop_threshold`, `call_rate_threshold`, and
`ask_budget_size` take effect after a daemon restart:

```bash
confire stop && confire start
```

### Analytics

| Key | Values | Default | Description |
|---|---|---|---|
| `analytics` | `true` / `false` | `true` | Send anonymous usage analytics |

### Advanced

| Key | Values | Description |
|---|---|---|
| `worker_url` | URL | Override the remote optimizer endpoint |

## Examples

```bash
# Reduce notification noise — only show large saves
confire config set notifications.min_saved_tokens=10000

# Use minimal notification style
confire config set notifications.style=minimal

# Turn off all notifications
confire config set notifications.enabled=false

# Set strict mode
confire config set mode=strict

# Opt out of analytics
confire config set analytics=false
```

## Per-agent overrides

To override capabilities for a specific agent, see
[Per-agent settings](../per-agent-settings).
