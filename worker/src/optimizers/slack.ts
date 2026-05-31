import type { InterceptEvent } from '../types.js'

const USER_ID_RE   = /\s*\(U[A-Z0-9]{8,11}\)/g
const MSG_TS_RE    = /\nMessage TS: \d+\.\d+/g
const FILE_META_RE = / \(ID: F[A-Z0-9]+, [^)]+\)/g
const PAGINATION_RE = /\n(?:There are no more messages[^\n]*|pagination_info[^\n]*)/g
const MENTION_RE   = /<@U[A-Z0-9]+\|([^>]+)>/g
const CHANNEL_RE   = /<#C[A-Z0-9]+\|([^>]+)>/g
const URL_PLAIN_RE = /<(https?:\/\/[^|>]+)>/g
const URL_LABELED_RE = /<(https?:\/\/[^|>]+)\|([^>]+)>/g
const MULTI_NL_RE  = /\n{3,}/g

function optimizeSlackText(text: string): string {
  let out = text
  out = out.replace(MENTION_RE, '@$1')
  out = out.replace(CHANNEL_RE, '#$1')
  out = out.replace(URL_LABELED_RE, '$2 ($1)')
  out = out.replace(URL_PLAIN_RE, '$1')
  out = out.replace(USER_ID_RE, '')
  out = out.replace(MSG_TS_RE, '')
  out = out.replace(FILE_META_RE, '')
  out = out.replace(PAGINATION_RE, '')
  out = out.replace(MULTI_NL_RE, '\n\n')
  return out.trim()
}

const MAX_MESSAGES = 30

export function optimizeSlack(rawText: string): string | null {
  if (!rawText || typeof rawText !== 'string') return null
  const isSlack = rawText.includes('Message TS:') || rawText.includes('=== THREAD') ||
    rawText.includes('--- Reply ') || (rawText.includes('From:') && rawText.includes('Reactions:'))
  if (!isSlack) return null
  let messageText = rawText
  try {
    const p = JSON.parse(rawText) as Record<string,string>
    if (p['messages']) messageText = p['messages']
  } catch { /* raw text */ }
  let text = optimizeSlackText(messageText)
  const msgCount = (text.match(/^From:/gm) ?? []).length
  if (msgCount > MAX_MESSAGES) {
    const msgs = text.split(/(?=^From:)/m)
    const kept = msgs.slice(-MAX_MESSAGES)
    text = `[confire: ${msgs.length - MAX_MESSAGES} older messages omitted]\n\n` + kept.join('')
  }
  return text
}

export function handlesSlack(event: InterceptEvent): boolean {
  return (event.tool?.name?.toLowerCase() ?? '').includes('slack')
}
