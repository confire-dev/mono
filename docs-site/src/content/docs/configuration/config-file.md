---
title: Config file
description: >-
  Confire local configuration — firewall mode, dashboard sync, policy
  sync, and notification settings.
---

Confire stores local configuration on your machine. You usually do not
need to edit the config file directly — use `confire config` instead.

```bash
confire config
confire config get mode
confire config set mode=strict
```

## Config location

By default, Confire stores local config at:

```
~/.confire/config.json
```

## Common commands

Show all config:

```bash
confire config
```

Read a setting:

```bash
confire config get mode
```

Update a setting:

```bash
confire config set mode=strict
```

## Firewall mode

| Key | Values | Default | Description |
|---|---|---|---|
| `mode` | `observe`, `balanced`, `strict`, `bypass` | `balanced` | Policy enforcement mode |

| Mode | Behavior |
|---|---|
| `observe` | Records matches without interrupting the agent |
| `balanced` | Default mode for daily work |
| `strict` | Reviews or blocks more aggressively |
| `bypass` | Temporarily disables enforcement |

```bash
confire config set mode=observe
confire config set mode=balanced
confire config set mode=strict
confire config set mode=bypass
```

You can also use:

```bash
confire mode observe
confire mode balanced
confire mode strict
confire mode bypass
```

## Firewall settings

| Key | Values | Default | Description |
|---|---|---|---|
| `firewall.enabled` | `true`, `false` | `true` | Master firewall toggle |
| `tool_firewall.enabled` | `true`, `false` | `true` | Enables PreToolUse evaluation |
| `tool_result_firewall.enabled` | `true`, `false` | `true` | Enables PostToolUse inspection |

Prefer changing mode instead of disabling the firewall entirely. Use
`observe` if you want Confire to record findings without interrupting
work.

## Dashboard sync

| Key | Values | Default | Description |
|---|---|---|---|
| `dashboard_sync.enabled` | `true`, `false` | `true` when logged in | Sync metadata-only security events to the dashboard |
| `dashboard_sync.include_raw_content` | `true`, `false` | `false` | Whether raw tool content may be synced |

Confire is local-first. Raw tool content should stay disabled by
default. Dashboard sync can include metadata such as rule ID, action
taken, risk level, tool category, client, timestamp, and session ID.

## Notifications

| Key | Values | Default | Description |
|---|---|---|---|
| `notifications.enabled` | `true`, `false` | `true` | Show Confire warnings and review messages |
| `notifications.style` | `brand`, `minimal` | `brand` | Notification format |

```bash
confire config set notifications.style=minimal
confire config set notifications.enabled=false
```

Do not disable notifications unless you know what you are doing.
Notifications are how Confire tells the agent and user about reviews,
warnings, and security findings.

## Policy sync

| Key | Values | Default | Description |
|---|---|---|---|
| `policy_sync.enabled` | `true`, `false` | `true` when logged in | Pull custom rules from Confire Cloud |
| `policy_sync.auto_pull` | `true`, `false` | `true` | Automatically refresh cached policy rules |

```bash
confire policy pull    # pull manually
confire policy status  # check policy state
```

## Security Registry

| Key | Values | Default | Description |
|---|---|---|---|
| `registry.enabled` | `true`, `false` | `true` | Pull signed security registry updates |
| `registry.auto_update` | `true`, `false` | `true` | Keep registry metadata updated automatically |

Registry updates are signed and evaluated locally.

## Advanced

| Key | Values | Description |
|---|---|---|
| `api_url` | URL | Override the Confire API endpoint |
| `log_level` | `debug`, `info`, `warn`, `error` | Controls local daemon logging |

```bash
confire config set log_level=debug
confire config set dashboard_sync.enabled=false
confire config set registry.auto_update=false
```

## Per-agent settings

Some settings can vary by client or project. See
[Per-agent settings](../per-agent-settings).
