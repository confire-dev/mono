---
title: Install
description: Get Confire installed and running in under a minute.
---

## Install the CLI

Run the installer:

```bash
curl -fsSL https://get.confire.dev/install.sh | sh
```

The installer:

- downloads the Confire binary for your platform,
- verifies the SHA-256 checksum,
- installs the binary to `~/.local/bin`,
- runs `confire setup`,
- installs supported agent hooks or gateway configuration,
- starts the local Confire daemon.

Confire supports macOS and Linux on x64 and arm64.

## Verify

```bash
confire version
confire status
```

`confire status` shows:

- which supported agents were detected,
- whether hooks or gateway configuration are installed,
- whether the daemon is running,
- which policy mode is active,
- whether the local firewall is ready.

A healthy install should show your agent connected and the daemon running.

## Next step

[Connect your agent →](../connect)
