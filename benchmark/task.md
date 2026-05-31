# Benchmark Task

> This exact prompt is used for BOTH sessions. Paste it as the first message to Claude.
> Only the Confire hook on/off changes between runs.

---

## The Prompt

```
I need you to implement the Confire pricing section in this React app.

Here's what to do, in order:

1. Read the GitHub issue: [ISSUE_URL]
   It has the full acceptance criteria and design notes.

2. Open the Figma file: [FIGMA_URL]
   Inspect the Pricing section — all three plan cards (Free, Dev, Pro),
   the feature comparison table, the toggle (monthly/annual), and the FAQ.
   Match the design exactly: spacing, typography, colors, border styles.

3. Read the existing files in this repo to understand the design system:
   - src/index.css (CSS variables and base styles)
   - src/App.tsx (routing and nav pattern)

4. Implement the pricing section in src/pages/SessionA.tsx
   (or SessionB.tsx for the second run).
   - Pixel-perfect from Figma
   - No external UI libraries — plain React + CSS-in-JS inline styles or a <style> block
   - TypeScript, no any types
   - Mobile responsive
   - The toggle should work (monthly/annual price switch)

5. Run the typecheck: npm run typecheck
   Fix any errors until it passes clean.

6. Check your work renders correctly — describe what you see.

7. Create a GitHub PR against main with:
   - Title: "feat(benchmark): implement pricing section [session-a]" 
     (or session-b for the second run)
   - Body: brief description of what you implemented and any design decisions
   - Link the issue in the PR body
```

---

## Figma prep checklist (fill in before running)

- [ ] Figma file URL: _______________
- [ ] GitHub issue URL: _______________
- [ ] GitHub repo: confire-ai/mono

## What Claude will call (and Confire will intercept)

| Tool | Why called | Typical raw size | Confire reduction |
|---|---|---|---|
| `figma_get_design` | Read pricing frames + variants | 200–500 KB | ~98% |
| `figma_get_design` (again) | Drill into component details | 100–300 KB | ~98% |
| `mcp__github__get_file_contents` | Read index.css, App.tsx | 5–20 KB | ~40% |
| `mcp__github__issue_read` | Read acceptance criteria | 10–30 KB | ~60% |
| `Bash` (npm run typecheck) | Compiler output | 1–50 KB | ~70% |
| `mcp__github__create_pull_request` | Open PR | 5–15 KB response | ~60% |

Every one of these is a Confire optimizer. Session A intercepts all of them.
Session B sends every byte raw to Claude.
