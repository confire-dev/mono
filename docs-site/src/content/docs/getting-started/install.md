---
title: Install
description: Get Confire installed and running in under a minute.
---

## Install the CLI

Run the installer:

```bash
curl -fsSL https://get.confire.dev/install.sh | sh
```

The installer downloads the binary for your platform (macOS or Linux,
x64 or arm64), verifies the SHA-256 checksum, places it in
`~/.local/bin`, runs `confire setup` to install agent hooks, and
starts the daemon automatically.

## Verify

```bash
confire version
confire status
```

`confire status` shows which agents were detected, whether hooks are
installed, and whether the daemon is running. If everything went
smoothly, you'll see green checkmarks next to your agent and the
context firewall status.

## Next step

[Connect your agent →](../connect)
