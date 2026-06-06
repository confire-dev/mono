import type { InterceptEvent } from '../types.js'

// Only these tools have optimization value — read_file / write_file / edit_file are intentional and untouched
const FS_OPT_TOOLS = new Set([
  'list_directory',
  'list_directory_with_sizes',
  'directory_tree',
  'get_file_info',
  'search_files',
  'list_allowed_directories',
])

// Default permissions that convey no signal
const DEFAULT_PERMS = new Set(['644', '755'])

function optimizeFileInfo(rawText: string): string | null {
  const lines = rawText.split('\n')
  const out: string[] = []
  let isDirVal: boolean | null = null

  for (const line of lines) {
    if (line.startsWith('accessed:')) continue // Almost never meaningful

    if (line.startsWith('isDirectory:')) {
      isDirVal = line.includes('true')
      continue
    }
    if (line.startsWith('isFile:')) {
      // Collapse the two booleans into a single "type" field
      if (isDirVal !== null) out.push(`type: ${isDirVal ? 'directory' : 'file'}`)
      isDirVal = null
      continue
    }
    if (line.startsWith('permissions:')) {
      const perm = line.split(':')[1]?.trim()
      if (perm && DEFAULT_PERMS.has(perm)) continue
    }
    out.push(line)
  }

  const result = out.join('\n').trim()
  return result !== rawText.trim() ? result : null
}

function optimizeDirectoryTree(rawText: string): string | null {
  let parsed: unknown
  try { parsed = JSON.parse(rawText) } catch { return null }

  const stripped = dropEmptyChildren(parsed)
  const out = JSON.stringify(stripped, null, 2)
  return out.length < rawText.length ? out : null
}

function dropEmptyChildren(node: unknown): unknown {
  if (Array.isArray(node)) return node.map(dropEmptyChildren)
  if (node && typeof node === 'object') {
    const o = node as Record<string, unknown>
    const result: Record<string, unknown> = {}
    for (const [k, v] of Object.entries(o)) {
      if (k === 'children' && Array.isArray(v) && v.length === 0) continue
      result[k] = dropEmptyChildren(v)
    }
    return result
  }
  return node
}

export function optimizeFilesystem(rawText: string, event: InterceptEvent): string | null {
  if (!rawText) return null
  const n = event.tool?.name ?? ''
  if (n === 'get_file_info')  return optimizeFileInfo(rawText)
  if (n === 'directory_tree') return optimizeDirectoryTree(rawText)
  // search_files and list_directory return paths/listings the model may need verbatim
  return null
}

export function handlesFilesystem(event: InterceptEvent): boolean {
  return FS_OPT_TOOLS.has(event.tool?.name ?? '')
}
