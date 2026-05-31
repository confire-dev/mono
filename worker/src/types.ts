// Shared envelope contract — mirrors the Go InterceptEvent / InterceptResult structs.
// This JSON shape is the ONLY thing shared between the Go daemon and the TS Worker.
// Change here and in confire/intercept/types.go together.

export type Phase =
  | 'session.start'
  | 'session.end'
  | 'tool.pre'
  | 'tool.post'
  | 'tool.batch.post'
  | 'turn.stop'
  | 'prompt.submit'
  | 'context.pre-compact' // disabled in v1 — slot reserved

export type Host = 'claude-code' | 'cursor' | 'cline' | 'windsurf' | 'codex' | string
export type Strategy = 'hooks' | 'mcp-proxy' | string

export interface InterceptSession {
  id: string
  cwd?: string
  transcriptPath?: string
}

export interface InterceptTool {
  name: string
  input?: unknown
  output?: unknown
  useId?: string
  isMcp: boolean
  mcpServer?: string
  durationMs?: number
}

export interface InterceptEvent {
  host: Host
  strategy: Strategy
  phase: Phase
  session: InterceptSession
  tool?: InterceptTool
  prompt?: string
  raw?: unknown
}

export type ResultKind =
  | 'passthrough'
  | 'replace-output'
  | 'replace-input'
  | 'add-context'
  | 'compact'

export interface InterceptStats {
  beforeBytes: number
  afterBytes: number
  optimizer: string
}

export interface InterceptResult {
  kind: ResultKind
  toolOutput?: unknown
  toolInput?: unknown
  context?: string
  stats?: InterceptStats
}

// ── Cloudflare Worker bindings ─────────────────────────────────────────────

export interface Env {
  // Cloudflare KV — content-hash optimizer cache + event idempotency
  CACHE: KVNamespace
  // Cloudflare Analytics Engine — real-time usage counters
  AE?: AnalyticsEngineDataset
  // Cloudflare Rate Limiting — per-user request throttling (one binding per plan tier)
  RL_FREE?: RateLimit
  RL_DEV?: RateLimit
  RL_PRO?: RateLimit
  // Supabase — billing truth, user data, dashboard
  SUPABASE_URL?: string
  SUPABASE_ANON_KEY?: string
  SUPABASE_SERVICE_KEY?: string
  // Amplitude — behavioral analytics only (never billing decisions)
  AMPLITUDE_KEY?: string
  // Stripe — checkout creation + webhooks
  STRIPE_SECRET_KEY?: string
  STRIPE_WEBHOOK_SECRET?: string
  // Supabase — database webhooks (set in wrangler secrets)
  SUPABASE_WEBHOOK_SECRET?: string
  // Environment tag
  ENVIRONMENT: string
  // Feature flags (set in wrangler.toml [vars] or via wrangler secret put)
  OPTIMIZER_API_ENABLED?: string  // "true" to enable POST /v1/optimize
}

// ── Worker wire formats ────────────────────────────────────────────────────

export interface SessionStartRequest {
  sessionId: string
  version: string
  host: Host
  strategies: Strategy[]
}

export interface SessionStartResponse {
  notification: string
  rulesVersion: string
  pullRules: boolean
}

export interface OptimizeRequest {
  event: InterceptEvent
  apiKey: string
}

export interface OptimizeResponse {
  result: InterceptResult
  // Optional warning sent to the daemon for display (80%/100% limit nudge, payload too large)
  _warning?: string
}

// ── Standalone Optimizer API (/v1/optimize) ────────────────────────────────
// General-purpose optimization endpoint — not tied to Claude Code hooks.
// Callers send any text/JSON before passing it to Groq, OpenAI, Gemini, etc.

export type OptimizerType =
  | 'auto'       // default — generic noise stripping
  | 'json'       // → generic optimizer
  | 'html'       // → webfetch optimizer
  | 'search'     // → websearch optimizer (Brave/Exa/Tavily JSON)
  | 'bash'       // → bash optimizer (log/test output)
  | 'github'     // → github optimizer
  | 'slack'      // → slack optimizer
  | 'figma'      // → figma optimizer
  | 'jira'       // → jira optimizer
  | 'notion'     // → notion optimizer
  | 'confluence' // → confluence optimizer
  | 'clickup'    // → clickup optimizer
  | 'amplitude'  // → amplitude optimizer
  | 'zapier'     // → zapier optimizer
  | 'playwright' // → playwright optimizer

export interface OptimizerApiRequest {
  content: string           // text, JSON, HTML, markdown — anything
  type?: OptimizerType      // hint for which optimizer to use (default: 'auto')
}

export interface OptimizerApiResponse {
  result: string            // optimized content (unchanged if no reduction found)
  optimized: boolean        // false = content returned as-is
  input_chars: number
  output_chars: number
  reduction_pct: number     // 0–100
  optimizer: string         // which optimizer ran
  cached: boolean
}
