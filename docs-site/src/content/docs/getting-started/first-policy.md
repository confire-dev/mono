---
title: See it working
description: Verify Confire is optimizing your context.
---

## Check status

```bash
confire status
```

This shows whether the daemon is running, which hooks are installed, and whether you're logged in.

## Watch it optimize

Run Claude Code and trigger a tool call — open a large file, run a shell command, or fetch a URL. You'll see Confire's optimization stats in the daemon output:

```bash
confire logs
```

## Check savings

After a session, run:

```bash
confire status
```

The output includes token savings for the current session.

## Log in for remote optimizers

Local optimizers (Bash, Read, WebFetch) work without an account. Remote optimizers — Figma, GitHub, Jira, Slack — require a Confire account:

```bash
confire login
```

This opens a browser to authenticate. Once logged in, remote optimizers activate automatically when the relevant tool types are detected.
