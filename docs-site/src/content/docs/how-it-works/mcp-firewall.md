---
title: MCP Firewall
description: Control which MCP servers and capabilities are reachable from your agent.
---

The MCP Firewall operates at a higher level than the Tool Firewall — it controls which MCP *servers* your agent can connect to, and which capability categories are exposed.

## Server-level allow/deny

```yaml
version: 1

mcp:
  servers:
    allow:
      - filesystem
      - github
    deny:
      - "*"   # deny everything not explicitly allowed
```

When `deny: ["*"]` is set, only servers listed under `allow` are reachable. Any attempt to connect to an unlisted server returns a connection error before any tools can be called.

## Capability filtering

MCP servers expose capabilities beyond just tools: resources, prompts, and sampling. You can restrict which capability types are available:

```yaml
mcp:
  capabilities:
    tools: allow
    resources: require-approval
    prompts: block
    sampling: block
```

This lets you use a server's tools while preventing the agent from requesting arbitrary file resources or injecting prompts.

## Tool category filtering

Rather than listing every tool by name, you can block by category:

```yaml
rules:
  - when:
      category: network      # any tool that makes outbound HTTP calls
    do: require-approval

  - when:
      category: destructive  # delete, overwrite, drop operations
    do: block
```

Confire maintains a built-in category map for common MCP servers. You can extend or override it in your policy file.

## Proxying vs. passthrough

By default Confire operates in *proxy mode*: all MCP traffic flows through it. If you have a server that shouldn't be touched (e.g., an internal tool with its own security controls), you can mark it as passthrough:

```yaml
mcp:
  servers:
    passthrough:
      - internal-db-server
```

Passthrough servers bypass the policy engine entirely — traffic goes directly from agent to server.

## Relationship to the Tool Firewall

The MCP Firewall and Tool Firewall work at different levels and complement each other:

| Layer           | Scope                              |
|-----------------|------------------------------------|
| MCP Firewall    | Which servers connect, which capabilities load |
| Tool Firewall   | Which individual tool calls execute |

A request must pass both layers. You can use the MCP Firewall to reduce attack surface broadly, then use the Tool Firewall to fine-tune specific call behavior.
