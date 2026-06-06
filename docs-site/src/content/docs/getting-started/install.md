---
title: Install the CLI
description: Get the Confire CLI installed in under a minute.
---

## Install

```bash
curl -fsSL https://get.confire.dev/install.sh | sh
```

The installer downloads the binary for your platform (macOS or Linux, x64 or arm64), verifies the SHA-256 checksum, places it in `~/.local/bin`, then runs `confire setup` and starts the daemon automatically.

## Verify

```bash
confire version
confire status
```

## Next step

[Connect Claude Code →](connect)
