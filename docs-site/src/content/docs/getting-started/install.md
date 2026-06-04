---
title: Install the CLI
description: Get the Confire CLI installed and running in under two minutes.
---

## Requirements

- Node.js 18 or later
- npm, pnpm, or Yarn
- A supported AI coding client (Claude Code, Cursor, or VS Code)

## Install

```bash
npm install -g @confire/cli
```

Or with pnpm:

```bash
pnpm add -g @confire/cli
```

Verify the install:

```bash
confire --version
```

## Authenticate

```bash
confire login
```

This opens a browser window to sign in with GitHub or email. Once authenticated, your credentials are stored locally in `~/.confire/credentials.json`.

If you prefer an API key (CI environments, headless servers):

```bash
confire login --api-key <your-key>
```

You can generate API keys in the [dashboard](https://app.confire.dev/dashboard/keys).

## Start the proxy

```bash
confire start
```

By default the proxy listens on `localhost:4747`. You can change the port:

```bash
confire start --port 9000
```

To run the proxy as a background daemon:

```bash
confire start --daemon
confire stop   # stop later
confire status # check if running
```

## Next step

[Connect your agent →](/getting-started/connect)
