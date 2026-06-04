---
title: VS Code
description: Connect VS Code Copilot to Confire via the HTTP MCP transport.
---

VS Code's GitHub Copilot extension supports MCP servers through workspace and user settings. Confire connects as an HTTP MCP server.

## Prerequisites

- Confire CLI installed (`npm install -g @confire/cli`)
- Confire proxy running (`confire start`)
- VS Code 1.99 or later with GitHub Copilot extension

## Configuration

**Workspace-level** (checked into the repo) — add to `.vscode/settings.json`:

```json
{
  "github.copilot.chat.mcp.servers": {
    "confire": {
      "url": "http://localhost:4747/mcp",
      "type": "http"
    }
  }
}
```

**User-level** (applies to all workspaces) — add to VS Code user settings:

```json
{
  "github.copilot.chat.mcp.servers": {
    "confire": {
      "url": "http://localhost:4747/mcp",
      "type": "http"
    }
  }
}
```

## Verify the connection

Open the GitHub Copilot Chat panel. In the tools picker (the `#` icon), you should see tools from your downstream MCP servers listed — these are proxied through Confire.

## Copilot agent mode

Confire's policy rules apply in both Copilot Chat and Copilot agent mode (`@workspace`). When the agent calls a tool, it flows through Confire's firewall before executing.

If a tool call is blocked, the agent sees a descriptive error message from Confire and can adapt its approach. If approval is required, a notification appears in VS Code.

## Notes

- VS Code MCP support is newer than Cursor's — check the Copilot extension changelog if you have connection issues.
- The proxy must be running before VS Code starts its MCP connection. If you start VS Code first, use the **GitHub Copilot: Refresh MCP Servers** command from the command palette.
- Project `.confire.yaml` files are applied automatically when that folder is the workspace root.
