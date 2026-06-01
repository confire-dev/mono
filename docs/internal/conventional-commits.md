# Conventional commits

All commits on `main` should follow [Conventional Commits](https://www.conventionalcommits.org/) so the bump-version workflow can semver, changelog, and release automatically.

## Format

```
<type>(<optional scope>): <short description>

[optional body]

[optional footer(s)]
```

Examples:

```
feat(platform): add Stripe top-up billing page
fix(cli): skip SessionStart hook when daemon is down
docs(internal): update deploy guide for R2 distribution
refactor(cli): load firewall rules from embedded JSON packs
chore(release): v0.3.0 [skip ci]
```

## Types

| Type | Semver bump | Changelog section |
|------|-------------|-------------------|
| `feat` | minor | Features |
| `feat!` or footer `BREAKING CHANGE:` | major | Features |
| `fix` | patch | Fixes |
| `docs`, `chore`, `refactor`, `perf`, `test`, `build`, `ci` | patch | Other |

Scopes are optional but encouraged (`cli`, `platform`, `worker`, `distribution`, etc.).

## Release automation

On every push to `main`, `.github/workflows/bump-version.yml`:

1. Reads commits since the last tag (or all history if no tags yet).
2. Chooses the next semver (`patch` / `minor` / `major`).
3. Prepends a grouped entry to `CHANGELOG.md`.
4. Commits `chore(release): vX.Y.Z [skip ci]`, tags, and pushes — which triggers `.github/workflows/release.yml`.

Do **not** hand-cut release tags unless you are bypassing automation. The release bump commit is ignored on the next run to avoid loops.

## Breaking changes

Use either:

```
feat(api)!: remove legacy /v0/optimize endpoint
```

or a footer:

```
feat(api): rename optimize response fields

BREAKING CHANGE: `tokensSaved` is now `tokens_saved`.
```
