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
  | 'sanitize'
  | 'block'
  | 'review'
  | 'warn'

export interface InterceptResult {
  kind: ResultKind
  toolOutput?: unknown
  toolInput?: unknown
  context?: string
  stats?: InterceptStats
}

export interface InterceptStats {
  beforeBytes: number
  afterBytes: number
  optimizer: string
}

// ── Cloudflare Worker bindings ─────────────────────────────────────────────

export interface Env {
  // Cloudflare KV — rules cache + event idempotency
  CACHE: KVNamespace
  // Cloudflare Analytics Engine — real-time usage counters
  AE?: AnalyticsEngineDataset
  // Cloudflare Rate Limiting
  RL_FREE?: RateLimit
  RL_DEV?: RateLimit
  RL_PRO?: RateLimit
  // Supabase — billing truth, user data, dashboard
  SUPABASE_URL?: string
  SUPABASE_ANON_KEY?: string
  SUPABASE_SERVICE_KEY?: string
  // Amplitude — behavioral analytics only
  AMPLITUDE_KEY?: string
  // Stripe — checkout creation + webhooks
  STRIPE_SECRET_KEY?: string
  STRIPE_WEBHOOK_SECRET?: string
  // Supabase — database webhooks
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

// ── Security event types ───────────────────────────────────────────────────

export type EventType =
  | 'PROMPT_INJECTION'
  | 'HIDDEN_TEXT'
  | 'SECRET_REDACTED'
  | 'CREDENTIAL_LURE'
  | 'DESTRUCTIVE_CMD'
  | 'SECRET_FILE_ACCESS'
  | 'MUTATING_MCP'
  | 'SOCIAL_ENGINEERING'

export type RiskLevel = 'NONE' | 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'

export type ActionTaken = 'ALLOWED' | 'WARNED' | 'REVIEW_REQUIRED' | 'BLOCKED' | 'SANITIZED'

export interface SecurityEvent {
  tool_name: string
  event_type: EventType
  risk_level: RiskLevel
  action_taken: ActionTaken
  pattern_matched?: string
  bypassed?: boolean
  session_id: string
}

export interface ProvenanceEvent {
  tool_name: string
  trust_level: 'internal' | 'external_trusted' | 'external_untrusted'
  sanitized: boolean
  redaction_count: number
  flags?: string[]
  session_id: string
}

export interface BypassEvent {
  tool_name?: string
  reason?: string
  session_id: string
}
