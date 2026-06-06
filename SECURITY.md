# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Confire, please **do not open a public GitHub issue**.

Report it privately to **hi@confire.dev** with the subject line `[SECURITY]`. Include:

- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept (if safe to share)
- Affected versions or components (CLI, worker, platform)

We will acknowledge your report within **2 business days** and aim to provide a resolution timeline within **7 days**. We'll keep you updated as we investigate and fix the issue.

## Scope

The following are in scope:

- `confire` CLI and daemon (`cli/`)
- Cloudflare Worker API (`worker/`)
- Platform web app (`platform/`)
- Authentication and API key handling
- Secret redaction and injection guard logic

Out of scope: third-party dependencies, the Supabase or Cloudflare infrastructure itself, and social engineering attacks.

## Data Security Model

Confire is designed with a local-first security posture:

- **Secret redaction runs locally** before any data leaves your machine. AWS keys, GitHub tokens, private keys, and other secrets are replaced with `[REDACTED:type]` before reaching the cloud worker or the AI model.
- **Prompt injection scanning runs locally.** Suspicious tool outputs are flagged or blocked on-device.
- **Your API key is stored in the OS keychain** (via system keyring), never in a plaintext config file.
- **Tool content only reaches the cloud worker if you are authenticated** and the remote optimizer is enabled. Without a valid API key, all optimization happens locally.

For a full breakdown of what data is sent remotely vs. kept local, see [PRIVACY.md](PRIVACY.md).

## Supported Versions

We support the latest released version of the CLI. Security fixes are not backported to older versions. Run `confire update` to stay current.
