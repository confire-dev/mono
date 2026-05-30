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
  // Supabase — billing truth, user data, dashboard
  SUPABASE_URL?: string
  SUPABASE_ANON_KEY?: string
  SUPABASE_SERVICE_KEY?: string
  // Amplitude — behavioral analytics only (never billing decisions)
  AMPLITUDE_KEY?: string
  // Stripe — webhooks
  STRIPE_WEBHOOK_SECRET?: string
  // Supabase — database webhooks (set in wrangler secrets)
  SUPABASE_WEBHOOK_SECRET?: string
  // Environment tag
  ENVIRONMENT: string
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
