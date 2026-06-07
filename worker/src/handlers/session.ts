import type { Env, SessionStartRequest, SessionStartResponse } from '../types.js'

const RULES_VERSION = '2025-05-30'

export async function handleSessionStart(request: Request, env: Env): Promise<Response> {
  let body: SessionStartRequest
  try {
    body = await request.json() as SessionStartRequest
  } catch {
    return Response.json({ error: 'invalid JSON' }, { status: 400 })
  }

  // TODO: look up usage stats from D1 for this API key
  // For now return a minimal notification
  const notification = buildNotification()
  const response: SessionStartResponse = {
    notification,
    rulesVersion: RULES_VERSION,
    pullRules: false,
  }
  return Response.json(response)
}

function buildNotification(): string {
  // One line max — confire is a firewall, it must not add noise.
  return `✓ Confire active · optimizing all tool outputs`
}
