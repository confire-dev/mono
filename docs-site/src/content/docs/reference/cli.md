---
title: CLI reference
description: All confire commands.
---

## `confire setup`

Installs Confire hooks into your AI agent's settings file.

```bash
confire setup           # interactive — asks for scope and agents
confire setup --global  # install into ~/.claude/settings.json
confire setup --local   # install into .claude/settings.json in nearest git root
```

After installation, restart your AI agent to activate the hook.

---

## `confire start`

Starts the optimizer daemon in the background.

```bash
confire start
```

The daemon listens on a Unix socket at `~/.confire/daemon.sock`. Setup runs this automatically.

---

## `confire stop`

Stops the running daemon.

```bash
confire stop
```

---

## `confire status`

Shows daemon status, hook installation, login state, and session stats.

```bash
confire status
```

---

## `confire login`

Opens a browser to authenticate with your Confire account. Stores the API key in the system keychain.

```bash
confire login
```

Required for remote optimizers (Figma, GitHub, etc.) and paid plan features.

---

## `confire logout`

Revokes the current API key and removes it from the keychain.

```bash
confire logout
```

---

## `confire reset`

Removes hooks from all agent settings files and stops the daemon. The binary stays installed.

```bash
confire reset
```

---

## `confire version`

Prints the installed version.

```bash
confire version
```

---

## `confire hook`

Called automatically by the Claude Code PostToolUse hook. Not intended for direct use.

```bash
confire hook  # reads tool output from stdin, writes optimized output to stdout
```

---

## Global flags

| Flag | Description |
|------|-------------|
| `--local` | Use local worker at `localhost:8787` |
| `--version` | Print version |
| `--help` | Show help |
