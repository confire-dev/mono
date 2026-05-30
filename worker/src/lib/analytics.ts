// Analytics helpers — writes to two destinations:
//   1. Cloudflare Analytics Engine  — real-time counters, dashboard-ready
//   2. Amplitude HTTP API           — user-level analytics, funnel analysis

export interface AnalyticsEvent {
  userId: string
  email: string
  eventType: string
  toolName?: string | undefined
  optimizer?: string | undefined
  beforeBytes?: number | undefined
  afterBytes?: number | undefined
  sessionId?: string | undefined
  host?: string | undefined
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

// writeAnalyticsEngine writes a data point to the Cloudflare Analytics Engine.
// The dataset binding (AE) is configured in wrangler.toml.
async function writeAnalyticsEngine(ae: AnalyticsEngineDataset, event: AnalyticsEvent): Promise<void> {
  const savedPct = event.beforeBytes && event.afterBytes
    ? Math.round((1 - event.afterBytes / event.beforeBytes) * 100)
    : 0

  ae.writeDataPoint({
    blobs: [
      event.userId,
      event.eventType,
      event.toolName ?? '',
      event.optimizer ?? '',
      event.host ?? 'claude-code',
      event.sessionId ?? '',
    ],
    doubles: [
      event.beforeBytes ?? 0,
      event.afterBytes ?? 0,
      savedPct,
    ],
    indexes: [event.userId],
  })
}

// sendAmplitude posts to Amplitude's HTTP API v2.
// Amplitude is used for user-level funnel analysis and retention.
async function sendAmplitude(apiKey: string, event: AnalyticsEvent): Promise<void> {
  const bytesSaved = (event.beforeBytes ?? 0) - (event.afterBytes ?? 0)
  await fetch('https://api2.amplitude.com/2/httpapi', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      api_key: apiKey,
      events: [{
        user_id: event.userId,
        event_type: event.eventType,
        time: Date.now(),
        event_properties: {
          tool_name:    event.toolName,
          optimizer:    event.optimizer,
          before_bytes: event.beforeBytes,
          after_bytes:  event.afterBytes,
          bytes_saved:  bytesSaved,
          session_id:   event.sessionId,
          host:         event.host ?? 'claude-code',
        },
        user_properties: { email: event.email },
      }],
    }),
  })
}
