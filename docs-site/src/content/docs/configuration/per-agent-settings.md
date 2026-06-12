---
title: Client settings
description: >-
  Confire supports Claude Code, Cursor, and VS Code. Setup is
  automatic. Most users do not need to configure clients manually.
---

Confire supports Claude Code, Cursor, and VS Code. You usually do not
need to configure each client manually. `confire setup` detects
supported clients and installs the right integration.

## Supported clients

| Client | Status |
|---|---|
| Claude Code | Supported |
| Cursor | Supported |
| VS Code | Supported |

## What Confire does across clients

Across supported clients, Confire can:

- review risky tool calls before they run,
- inspect tool results after they return,
- add firewall context where supported,
- record local security events,
- sync metadata-only events when dashboard sync is enabled.

Exact behavior can vary by client and tool surface, but the product
behavior is the same: Confire gives your agent a local firewall around
tool activity.

## Enable or disable a client

Run setup again to change which clients are connected:

```bash
confire setup
```

Inspect detected clients:

```bash
confire status
```

## Client IDs

Confire uses these client IDs internally:

| Client | ID |
|---|---|
| Claude Code | `claude-code` |
| Cursor | `cursor` |
| VS Code | `vscode` |

## Advanced overrides

Most users should not need client-specific overrides. If needed,
you can disable a client integration from config:

```bash
confire config set clients.vscode.enabled=false
```

Re-enable it:

```bash
confire config set clients.vscode.enabled=true
```

Then run setup again:

```bash
confire setup
```

## Check effective setup

```bash
confire status
```

The **Agents** section shows which clients were detected and whether
Confire is connected.
