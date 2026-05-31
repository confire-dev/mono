# Confire Benchmark

Same task. Same design. Same repo. Confire on vs off.

## Setup

```bash
cd benchmark
pnpm install
pnpm dev        # http://localhost:5173
```

## Running the benchmark

### Session A — Confire ON

```bash
confire start
# verify it's running:
confire status
```

Open Claude Code in this repo. Paste the prompt from `task.md` (fill in the Figma + issue URLs first).

After Claude finishes:
```bash
confire stats --today   # tokens saved, per-tool breakdown
# also check /cost in Claude Code for session spend
```

Record results in RESULTS.md → Session A column.

---

### Session B — Confire OFF

```bash
confire stop
# verify it's off:
confire status
```

Open a **fresh** Claude Code session in this repo. Paste the **exact same prompt** from `task.md`, but target `SessionB.tsx`.

After Claude finishes:
```bash
# check /cost in Claude Code for session spend
```

Record results in RESULTS.md → Session B column.

---

## What to record

See `RESULTS.md` for the template.

## Why this task

The pricing section task was chosen to hit every Confire optimizer:

- **Figma** — multiple `figma_get_design` calls on a complex multi-frame design (biggest win, ~98%)
- **GitHub issue** — reads verbose JSON with labels, comments, metadata (~60%)  
- **Read files** — reads existing source files for context (~40%)
- **Bash** — typecheck output, potentially noisy compiler errors (~70%)
- **GitHub PR** — creates PR, reads back the response (~60%)

A simple "write me a button" task wouldn't show much.
This task guarantees multiple large tool calls across every optimizer type.
