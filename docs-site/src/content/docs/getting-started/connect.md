---
title: Connect your agents
description: Point Claude Code, Cursor, or VS Code at the Confire proxy.
---

Once the proxy is running (`confire start`), tell your agent to route its MCP traffic through it.

## Claude Code

Add Confire as an MCP server in your Claude Code config (`.claude/settings.json` or `~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "confire": {
      "command": "confire",
      "args": ["mcp"]
    }
  }
}
```

The `confire mcp` sub-command starts an MCP stdio transport that forwards to the proxy.

For a detailed walk-through see the [Claude Code client guide →](/clients/claude-code)

## Cursor

In Cursor settings, add a new MCP server pointing at the Confire HTTP proxy:

```json
{
  "mcpServers": {
    "confire": {
      "url": "http://localhost:4747/mcp"
    }
  }
}
```

For a detailed walk-through see the [Cursor client guide →](/clients/cursor)

## VS Code (Copilot)

In your workspace `.vscode/settings.json`:

```json
{
  "github.copilot.mcpServers": {
    "confire": {
      "url": "http://localhost:4747/mcp"
    }
  }
}
```

For a detailed walk-through see the [VS Code client guide →](/clients/vscode)

## Verify the connection

After connecting, run a quick sanity check:

```bash
confire test
```

This sends a test request through the proxy and prints what the firewall would do with it.

## Next step

[Test your first policy →](/getting-started/first-policy)
