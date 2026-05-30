import type { InterceptEvent } from '../types.js'

const MIME_SHORT: Record<string,string> = {
  'application/vnd.google-apps.spreadsheet': 'sheet',
  'application/vnd.google-apps.document': 'doc',
  'application/vnd.google-apps.presentation': 'slides',
  'application/vnd.google-apps.folder': 'folder',
  'application/vnd.google-apps.form': 'form',
  'application/pdf': 'pdf', 'text/plain': 'txt', 'image/png': 'png', 'image/jpeg': 'jpg',
}

function shortMime(mime: string): string {
  return MIME_SHORT[mime] ?? mime.split('/').pop()!.replace('vnd.', '').slice(0, 12)
}
function humanSize(bytes: unknown): string {
  const n = parseInt(String(bytes ?? 0))
  if (n > 1_000_000) return `${Math.round(n / 1_000_000)}MB`
  if (n > 1_000) return `${Math.round(n / 1_000)}KB`
  return `${n}B`
}
function cleanFile(f: Record<string,unknown>): Record<string,unknown> {
  return {
    id: f['id'], title: f['title'] ?? f['name'],
    type: shortMime(f['mimeType'] as string),
    size: humanSize(f['fileSize'] ?? f['size']),
    modified: ((f['modifiedTime'] ?? f['modifiedByMeTime'] ?? '') as string)?.slice(0, 10),
    owner: f['owner'] ?? (f['owners'] as Record<string,string>[])?.[0]?.['emailAddress'],
    url: f['viewUrl'] ?? f['webViewLink'],
  }
}

export function optimizeGoogleDrive(rawText: string): string | null {
  if (!rawText) return null
  let data: unknown
  try { data = JSON.parse(rawText) } catch { return null }
  if (!data || typeof data !== 'object') return null
  const d = data as Record<string,unknown>

  if (Array.isArray(d['files'])) {
    return JSON.stringify({
      files: (d['files'] as Record<string,unknown>[]).map(cleanFile),
      ...(d['nextPageToken'] ? { next_page: '[cursor]' } : {}),
    }, null, 2)
  }
  if (d['id'] && d['mimeType'] && !d['files']) {
    const f = d as Record<string,unknown>
    return JSON.stringify({
      id: f['id'], title: f['title'] ?? f['name'], type: shortMime(f['mimeType'] as string),
      size: humanSize(f['size'] ?? f['fileSize']),
      modified: (f['modifiedTime'] as string)?.slice(0, 10),
      created: (f['createdTime'] as string)?.slice(0, 10),
      owner: f['owner'] ?? (f['owners'] as Record<string,string>[])?.[0]?.['emailAddress'],
      url: f['viewUrl'] ?? f['webViewLink'],
      description: f['description'] || undefined,
    }, null, 2)
  }
  return null
}

export function handlesGoogleDrive(event: InterceptEvent): boolean {
  const n = event.tool?.name?.toLowerCase() ?? ''
  return n.includes('google_drive') || n.includes('googledrive')
}
