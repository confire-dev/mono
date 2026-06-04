---
title: Claude Code
description: Connect Claude Code to Confire via an MCP server.
---

Claude Code connects to Confire using its built-in MCP server support. Once connected, all tool calls Claude Code makes pass through the Confire proxy before reaching the underlying MCP servers.

## Prerequisites

- Confire CLI installed (`npm install -g @confire/cli`)
- Confire proxy running (`confire start`)
- Claude Code 1.0 or later

## Configuration

Add Confire as a local MCP server in your Claude Code settings.

**Global** (applies to all projects) — edit `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "confire": {
      "command": "confire",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

**Project-level** (checked into the repo) — edit `.claude/settings.json` in your project root:

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

The project-level config takes precedence over global when both exist.

## How it works

When you add Confire as an MCP server, Claude Code spawns a `confire mcp` process. This process starts an MCP stdio transport that:

1. Advertises all the tools from your configured downstream MCP servers
2. Intercepts every `tools/call` before it reaches the real server
3. Applies your policy rules (firewall, optimizer, redaction, injection guard)
4. Forwards allowed calls to the real server and returns the result

From Claude Code's perspective, it's talking to a single MCP server that happens to expose all your other tools.

## Verify the connection

In Claude Code, run:

```
/mcp
```

You should see `confire` listed with a green status and the tool count matching your downstream servers.

Or from the CLI:

```bash
confire test --client claude-code
```

## Policy file for Claude Code

Put a `.confire.yaml` in your project root to apply project-specific rules:

```yaml
version: 1

rules:
  # Always require approval before writing to sensitive directories
  - when:
      tool: write_file
      args:
        path: "^/(etc|usr|bin|sbin)"
    do: block

  # Approve shell commands
  - when:
      tool: bash
    do: require-approval
```

Claude Code respects this file automatically — no restart required.

## Hooks integration

If you use Claude Code hooks (`.claude/hooks.json`), Confire works alongside them. Hooks run after Confire's policy engine, so your hooks will only see tool calls that Confire allowed.
