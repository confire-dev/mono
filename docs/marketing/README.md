# Confire Marketing Docs

Generated from [coreyhaines31/marketingskills](https://github.com/coreyhaines31/marketingskills) global skills (`~/.agents/skills/`). Each file follows that skill's output format, applied to Confire using [product-context.md](./product-context.md).

**Commit note:** SessionStart silence fix committed separately (`fix(cli): stop SessionStart status banner on healthy sessions`).

---

## Documents

| Skill | File | What it covers |
|-------|------|----------------|
| — | [product-context.md](./product-context.md) | Positioning, ICP, free vs paid, voice |
| ai-seo | [ai-seo.md](./ai-seo.md) | AI search visibility, citation strategy, llms.txt |
| community-marketing | [community-marketing.md](./community-marketing.md) | Discord/GitHub community, ambassador program |
| content-strategy | [content-strategy.md](./content-strategy.md) | Pillars, topic clusters, 90-day editorial calendar |
| copy-editing | [copy-editing.md](./copy-editing.md) | Audit of existing site copy + refresh recommendations |
| launch | [launch.md](./launch.md) | v1 GA launch plan (Claude Code + Cursor + VS Code) |
| marketing-ideas | [marketing-ideas.md](./marketing-ideas.md) | Top tactics for current stage, prioritized |
| marketing-plan | [marketing-plan.md](./marketing-plan.md) | 12-month AARRR plan (bootstrapped) |
| marketing-psychology | [marketing-psychology.md](./marketing-psychology.md) | Principles applied to Confire pages and CLI |
| onboarding | [onboarding.md](./onboarding.md) | Signup → setup → first savings activation flow |
| site-architecture | [site-architecture.md](./site-architecture.md) | confire.dev IA, nav, URLs, internal linking |

---

## How to use

1. Read **product-context.md** first — all other docs assume it.
2. **site-architecture** + **content-strategy** → what to build on the site.
3. **copy-editing** → refresh existing copy in `docs/copy/` before shipping.
4. **launch** + **marketing-plan** → sequencing for GA.
5. **ai-seo** → structure new pages for Google AI Overviews + ChatGPT/Perplexity citations.

## Related repo copy

Existing draft copy lives in `docs/copy/` (homepage, pricing, features, CTAs, onboarding, SEO). This folder is the strategic layer; `docs/copy/` is the execution layer.
