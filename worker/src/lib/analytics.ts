// Analytics helpers — writes to two destinations:
//   1. Cloudflare Analytics Engine  — real-time counters, dashboard-ready
//   2. Amplitude HTTP API           — user-level analytics, funnel analysis

export interface AnalyticsEvent {
  userId: string
  email: string
  eventType: string
  toolName?: string | undefined
  riskLevel?: string | undefined
  actionTaken?: string | undefined
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

async function writeAnalyticsEngine(ae: AnalyticsEngineDataset, event: AnalyticsEvent): Promise<void> {
  ae.writeDataPoint({
    blobs: [
      event.userId,
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
        user_id: event.userId,
        event_type: event.eventType,
        time: Date.now(),
        event_properties: {
          tool_name:    event.toolName,
          risk_level:   event.riskLevel,
          action_taken: event.actionTaken,
          session_id:   event.sessionId,
          host:         event.host ?? 'claude-code',
        },
        user_properties: { email: event.email },
      }],
    }),
  })
}
