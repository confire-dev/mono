---
title: Remote policy sync
description: Keep policies consistent across machines and team members using remote sync.
---

Remote policy sync lets you store your policy in the Confire cloud and pull it down automatically on any machine. This is useful for teams that want a shared baseline policy, or for keeping your personal policy in sync across multiple computers.

## How it works

1. You push a policy to the Confire cloud with `confire policy push`
2. The cloud stores a versioned copy tied to your account
3. When the local proxy starts, it fetches the latest remote policy
4. Local project-level `.confire.yaml` files still override the remote policy

## Setup

### Push your policy

```bash
confire policy push ~/.confire/policy.yaml
```

Or push a named policy (useful for teams with multiple environments):

```bash
confire policy push ./staging-policy.yaml --name staging
confire policy push ./prod-policy.yaml --name production
```

### Pull on another machine

```bash
confire policy pull          # pulls the default policy
confire policy pull --name staging
```

### Enable auto-sync

To have the proxy always fetch the latest on start:

```yaml
# ~/.confire/config.yaml
sync:
  enabled: true
  policy: default     # or a named policy
  interval: 1h        # how often to check for updates (default: on start)
```

## Team policies

On a Dev or Pro plan, you can share a named policy with your team. Anyone on your team account can pull it:

```bash
# Admin: push shared policy
confire policy push ./team-policy.yaml --name team-baseline --share

# Team member: pull it
confire policy pull --name team-baseline --account your-org
```

## Policy versioning

Every push creates a new version. You can list and roll back:

```bash
confire policy versions
# v5  2025-01-15  10:23  (current)
# v4  2025-01-14  09:11
# v3  2025-01-10  14:55

confire policy rollback --version 4
```

## Merge behavior

When both a remote policy and a local `.confire.yaml` exist, they are merged:

1. Local project rules are evaluated first
2. Remote (pulled) global rules are evaluated second
3. The `default` action from the local config wins if present

This means you can use the remote policy as a company-wide baseline and let teams customize per-project without overriding everything.

## Requirements

Remote policy sync requires a **Dev plan or above**. On the Free plan, policies are local-only.
