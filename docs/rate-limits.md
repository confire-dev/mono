# Rate Limits

Confire enforces two types of limits to ensure fair access for all users.

---

## Per-minute request rate limits

These limits apply to calls to the remote optimizer API (`POST /api/optimize`).
The local CLI optimizer is unlimited and is never counted or rate-limited.

| Plan       | Requests per minute |
|------------|---------------------|
| Free       | 20                  |
| Dev        | 60                  |
| Pro        | 200                 |
| Enterprise | 200 (customizable)  |

When a rate limit is exceeded the API returns `HTTP 429` with a `Retry-After: 60` header.
The CLI will display a warning and fall back to local optimization automatically.

**In practice:** A typical Claude Code agentic session fires 1–5 tool calls per second at peak.
Even at 5 calls/second, Free users would need a 4-second burst to hit 20/minute,
and Pro users would need a sustained 3.3 calls/second for a full minute to hit 200.
Normal agent usage stays well below these limits.

---

## Monthly token volume (Pro "unlimited fair-use")

Pro and Pro Annual plans have no optimization count limit, but they do have a monthly token cap.

| Plan              | Remote tokens per month |
|-------------------|-------------------------|
| Free              | 5 million               |
| Dev / Dev Annual  | 50 million              |
| Pro / Pro Annual  | 500 million             |
| Enterprise        | Custom                  |

Tokens are estimated from payload size (`bytes / 4`). A typical tool response
is 1–50 KB, so 500 million tokens ≈ roughly 10,000–500,000 optimizations per month.

When you approach 80% of your monthly token budget, the CLI shows a warning nudge.
When the budget is exhausted, remote optimization is paused for the rest of the period
and the CLI falls back to local optimization.

---

## Per-call payload limits

Individual tool responses above a certain size are not sent to the remote optimizer
(they skip directly to local optimization instead).

| Plan              | Max payload per call |
|-------------------|----------------------|
| Free              | 2 MB                 |
| Dev / Dev Annual  | 5 MB                 |
| Pro / Pro Annual  | 10 MB                |
| Enterprise        | 50 MB                |

The absolute Worker-level hard cap is 50 MB regardless of plan.

---

## Local optimizer

The local optimizer (embedded in the CLI binary) is **unlimited** — no rate limits,
no monthly quotas, no payload caps. It runs on your machine at process speed.
Local optimization is always available, even when offline or when remote limits are reached.

---

## Errors and fallback behavior

| Error              | HTTP status | CLI behavior                            |
|--------------------|-------------|-----------------------------------------|
| Rate limit hit     | 429         | Falls back to local optimizer, retries next call |
| Monthly quota full | 200 (soft)  | `add-context` result, warning in session |
| Payload too large  | 200 (soft)  | `passthrough` result, local optimizer runs |
| API unreachable    | —           | Falls back to local optimizer silently   |

The CLI never blocks a tool call due to Confire errors — worst case it passes the output through unchanged.
