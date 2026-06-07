---
title: CLI reference
description: All confire commands and global flags.
---

## `confire setup`

Installs Confire hooks into detected AI agent settings files.

```bash
confire setup            # interactive — prompts for scope and agents
confire setup --global   # install into user-level settings
confire setup --local    # install into nearest git root's settings
```

Restart your agent after running setup.

---

## `confire start`

Starts the optimizer daemon in the background.

```bash
confire start
```

The daemon listens on a Unix socket at `~/.confire/daemon.sock`
and handles all firewall evaluation and optimization. Setup starts
it automatically.

---

## `confire stop`

Stops the running daemon.

```bash
confire stop
```

---

## `confire status`

Shows daemon state, hook installation, account info, and
optimizer status.

```bash
confire status
```

---

## `confire on`

Enables the firewall and sets the mode to `balanced`.

```bash
confire on
```

---

## `confire off`

Sets the mode to `bypass`, disabling all firewall enforcement.

```bash
confire off
```

---

## `confire bypass-next`

Sets a one-shot flag to allow the next PreToolUse event to skip
firewall review. Clears automatically after one use.

```bash
confire bypass-next
```

---

## `confire login`

Opens a browser to authenticate with your Confire account.
Stores the API key in the system keychain.

```bash
confire login
```

Required for remote optimizers (Figma, GitHub, all MCP tools)
and paid plan features.

---

## `confire logout`

Revokes the current API key and removes it from the keychain.

```bash
confire logout
```

---

## `confire policy`

Subcommands for managing firewall policy.

```bash
confire policy status                     # show mode, rule counts, cache info
confire policy test <command-or-tool>     # simulate a PreToolUse evaluation
confire policy pull                       # fetch custom rules (paid plan)
```

Examples:

```bash
confire policy test 'git push --force'
confire policy test 'mcp__github__merge_pull_request'
confire policy test 'git log'
```

---

## `confire config`

Read or write CLI configuration stored in
`~/.confire/config.json`.

```bash
confire config                              # show all settings
confire config get notifications.style
confire config set notifications.enabled=false
confire config set mode=strict
```

See [Config file](../../configuration/config-file) for all keys.

---

## `confire reset`

Removes hooks from all agent settings files and stops the daemon.
The binary stays installed.

```bash
confire reset
```

---

## `confire update`

Updates the Confire binary to the latest release.

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

Called automatically by agent hooks. Not for direct use.

```bash
confire hook   # reads tool event from stdin, writes result to stdout
```

---

## Global flags

| Flag | Description |
|---|---|
| `--local` | Use local dev servers (platform `:4321`, worker `:8787`) |
| `--version` | Print version and exit |
| `--help` | Show help |
