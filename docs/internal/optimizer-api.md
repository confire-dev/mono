# Context Optimization API — Product Concept & Architecture

> **Status: not yet launched.**
> The endpoint exists in the Worker but returns 404 until `OPTIMIZER_API_ENABLED="true"` is set.
> See [deploy.md](./deploy.md#feature-flags) for how to enable it.

## What it is

A standalone HTTP endpoint (`POST /v1/optimize`) that reduces any text or JSON payload
before the caller sends it to an LLM. Same optimizer engine, same auth, same billing —
exposed as a general-purpose pre-LLM context reduction API.

```
User builds context (RAG docs, tool outputs, conversation history)
    ↓
POST api.confire.dev/v1/optimize  { content: "...", type: "json" }
    ↓  ~5ms round-trip
{ result: "...", reduction_pct: 42, optimizer: "generic" }
    ↓
POST api.groq.com  ← 42% fewer tokens, 42% cheaper on that call
```

**Selling point:** "Call us before calling Groq. Pay us once, pay the LLM less every time."

---

## Architecture

The endpoint is a thin adapter over the exact same optimizer engine used by the Claude Code
hook path. No optimizer code was changed to support this — the adaptation is:

```
content + type hint
    ↓
TYPE_TO_TOOL map  →  tool name (e.g. "json" → "generic", "html" → "WebFetch")
    ↓
synthetic InterceptEvent  { host: "optimizer-api", phase: "tool.post", tool.output: content }
    ↓
handle(event)  ←  SAME function as /optimize endpoint
    ↓
result.toolOutput  →  OptimizerApiResponse { result, reduction_pct, optimizer, ... }
```

All infrastructure is shared:
- Auth (`/lib/auth.ts`)
- Rate limiting (`RL_FREE` / `RL_DEV` / `RL_PRO` bindings)
- Entitlement check (`canUseRemoteOptimizer`)
- Content-hash cache (KV)
- Usage accounting (`recordOptimization`)
- Analytics (`trackEvent`)

---

## Request / Response

**Request** (`POST /v1/optimize`):

```json
{
  "content": "<text, JSON, HTML, or any string>",
  "type": "auto"
}
```

`type` is optional (default `auto`). Valid values:

| type        | optimizer used      | best for                        |
|-------------|---------------------|---------------------------------|
| `auto`      | generic             | unknown / mixed content         |
| `json`      | generic             | API responses, structured data  |
| `html`      | webfetch            | web pages, scraped content      |
| `search`    | websearch           | Brave/Exa/Tavily results JSON   |
| `bash`      | bash                | shell output, test logs         |
| `github`    | github              | PR objects, diff text           |
| `slack`     | slack               | Slack thread exports            |
| `figma`     | figma               | Figma file JSON                 |
| `jira`      | atlassian           | Jira issue JSON                 |
| `notion`    | notion              | Notion block/page JSON          |
| `confluence`| atlassian           | Confluence page content         |
| `clickup`   | clickup             | ClickUp task/comment data       |
| `amplitude` | amplitude           | Amplitude analytics exports     |
| `zapier`    | zapier              | Zapier webhook payloads         |
| `playwright`| playwright          | Playwright test output / traces |

**Response:**

```json
{
  "result": "<optimized content>",
  "optimized": true,
  "input_chars": 8200,
  "output_chars": 4100,
  "reduction_pct": 50,
  "optimizer": "generic",
  "cached": false
}
```

`optimized: false` means no reduction was found — `result` equals the original `content`.

When approaching the monthly quota limit, the response includes `X-Confire-Warning` header.

---

## Auth

Same API key used for the Claude Code hook. Pass as `Authorization: Bearer <key>`.

```bash
curl https://api.confire.dev/v1/optimize \
  -H "Authorization: Bearer cfr_..." \
  -H "Content-Type: application/json" \
  -d '{"content": "{\"node_id\":\"MDQ6\",\"avatar_url\":\"https://...\",\"title\":\"fix bug\",\"state\":\"open\"}", "type": "json"}'
```

---

## Quotas and rate limits

Shared with the Claude Code hook quota (same `cloudOptimizationsMonthly` counter in v1).

| Plan       | Requests/month   | Requests/minute |
|------------|------------------|-----------------|
| Free       | 500              | 20              |
| Dev        | 5,000            | 60              |
| Pro        | Unlimited (fair-use) | 200         |
| Enterprise | Custom           | 200+            |

---

## Error responses

| Status | Error                | Meaning                                   |
|--------|----------------------|-------------------------------------------|
| 400    | `invalid JSON`       | Request body is not valid JSON            |
| 400    | `missing or empty content` | `content` field missing or empty   |
| 401    | `missing Authorization header` | No API key                    |
| 401    | `invalid API key`    | Key not found or revoked                  |
| 402    | `entitlement_denied` | Plan limit reached or optimizer not included |
| 413    | `payload_too_large`  | `content` exceeds plan or absolute limit  |
| 429    | `rate_limit_exceeded`| Too many requests this minute             |

---

## Future considerations

- **Separate quota** (`apiOptimizationsMonthly`): if the API use case grows, give it its own
  limit so Claude Code hook users and standalone API users don't compete for the same budget.
- **Streaming support**: not needed — optimizer is sync, ~5ms. No benefit to streaming.
- **SDK packages**: `@confire/node`, `@confire/python` wrappers around the endpoint.
- **Batch endpoint** (`POST /v1/optimize/batch`): optimize multiple items in one call.
  Maps cleanly to `Promise.all(items.map(item => dispatch(item)))`.
