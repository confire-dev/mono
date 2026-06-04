---
title: Cursor
description: Connect Cursor to Confire via the HTTP MCP transport.
---

Cursor supports MCP servers over HTTP. Confire's proxy exposes an HTTP MCP endpoint that Cursor connects to directly.

## Prerequisites

- Confire CLI installed (`npm install -g @confire/cli`)
- Confire proxy running (`confire start`)
- Cursor 0.46 or later (MCP support required)

## Configuration

Open Cursor settings and navigate to **Features → MCP Servers**, or edit the config file directly.

**Via Cursor UI:**

1. Open Settings (`⌘,` or `Ctrl+,`)
2. Search for "MCP"
3. Click **Add MCP Server**
4. Set type to **HTTP**
5. Set URL to `http://localhost:4747/mcp`
6. Name it `confire`

**Via config file** (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "confire": {
      "url": "http://localhost:4747/mcp",
      "transport": "http"
    }
  }
}
```

## Custom port

If you're running the proxy on a non-default port:

```bash
confire start --port 9000
```

Update the URL in Cursor settings accordingly:

```json
{
  "mcpServers": {
    "confire": {
      "url": "http://localhost:9000/mcp"
    }
  }
}
```

## Verify the connection

In Cursor, open the MCP panel. You should see `confire` listed with a connected status and the tool list populated.

From the CLI:

```bash
confire test --client cursor
```

## Notes

- Cursor uses SSE (Server-Sent Events) for streaming tool results. Confire supports this natively.
- Project-level policy files (`.confire.yaml` in your project root) are picked up automatically when you open that folder in Cursor.
- Cursor's Composer and Chat both route MCP calls through the same connection, so Confire's rules apply to both.
