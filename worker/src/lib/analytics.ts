// Analytics helpers — writes to two destinations:
//   1. Cloudflare Analytics Engine  — real-time counters, dashboard-ready
//   2. Amplitude HTTP API           — user-level analytics, funnel analysis

export interface AnalyticsEvent {
  userId:         string   // DB UUID — never email
  eventType:      string
  // Firewall / security
  decision?:      string    // block | review | warn | sanitize
  toolName?:      string
  riskLevel?:     string
  actionTaken?:   string
  // Session context
  sessionId?:     string
  host?:          string
  // Provenance categories (never raw MCP server names or domains)
  mcpServerKnown?:    boolean
  mcpServerCategory?: string
  originCategory?:    string
  originKnownPublic?: boolean
  // Meta
  source?:         string   // cli | dashboard
  confireVersion?: string
}

// trackEvent writes to both Analytics Engine and Amplitude.
// Fire-and-forget: call without await to avoid blocking the response.
export function trackEvent(
  ae: AnalyticsEngineDataset | undefined,
  amplitudeKey: string | undefined,
  event: AnalyticsEvent,
): void {
  if (ae) writeAnalyticsEngine(ae, event).catch(() => {})
  if (amplitudeKey) sendAmplitude(amplitudeKey, event).catch(() => {})
}

async function writeAnalyticsEngine(ae: AnalyticsEngineDataset, event: AnalyticsEvent): Promise<void> {
  ae.writeDataPoint({
    blobs: [
      event.userId,   // DB UUID
      event.eventType,
      event.toolName ?? '',
      event.riskLevel ?? '',
      event.host ?? 'claude-code',
      event.sessionId ?? '',
    ],
    doubles: [],
    indexes: [event.userId],
  })
}

async function sendAmplitude(apiKey: string, event: AnalyticsEvent): Promise<void> {
  await fetch('https://api2.amplitude.com/2/httpapi', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      api_key: apiKey,
      events: [{
        user_id:    event.userId,
        event_type: event.eventType,
        time:       Date.now(),
        event_properties: {
          ...(event.decision          ? { decision:            event.decision }          : {}),
          ...(event.toolName          ? { tool_category:       event.toolName }          : {}),
          ...(event.riskLevel         ? { risk_level:          event.riskLevel }         : {}),
          ...(event.host              ? { host:                event.host }              : {}),
          ...(event.mcpServerKnown    !== undefined ? { mcp_server_known:     event.mcpServerKnown }    : {}),
          ...(event.mcpServerCategory ? { mcp_server_category: event.mcpServerCategory } : {}),
          ...(event.originCategory    ? { origin_category:     event.originCategory }    : {}),
          ...(event.originKnownPublic !== undefined ? { origin_known_public:  event.originKnownPublic } : {}),
          ...(event.source            ? { source:              event.source }            : {}),
          ...(event.confireVersion    ? { confire_version:     event.confireVersion }    : {}),
        },
      }],
    }),
  })
}
