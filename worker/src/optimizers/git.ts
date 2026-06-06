import type { InterceptEvent } from '../types.js'

// Tools this optimizer handles (read/write ops return tiny success strings — no value in optimizing them)
const GIT_OPT_TOOLS = new Set(['git_status', 'git_diff', 'git_diff_unstaged', 'git_diff_staged', 'git_log', 'git_show'])

// Only machine-generated internals that are never human-readable —
// lock files and dist/ are kept because the model may be answering
// questions like "what packages changed?" or "what was the build output?"
const NOISE_PATH_RE = /^node_modules\//

function isNoisyFile(filePath: string): boolean {
  if (NOISE_PATH_RE.test(filePath)) return true
  // Minified bundles — no human-readable signal
  if (/\.min\.[jt]sx?$/.test(filePath) || filePath.endsWith('.min.css')) return true
  return false
}

// git.Actor repr object → human name
const GIT_ACTOR_RE = /<git\.Actor "([^"]+)" <([^>]+)>>/g
// repr-quoted strings on Commit / Message lines
const REPR_LINE_RE  = /^((?:Commit|Message|Author)): '((?:[^'\\]|\\.)*)'\s*$/gm
// git status hint lines — lines containing (use "cmd ...")
const STATUS_HINT_RE = /^[ \t]+\(use "[^"]+"\)[^\n]*/gm

function cleanActors(s: string): string {
  return s.replace(GIT_ACTOR_RE, '$1 <$2>')
}

function cleanGitLog(rawText: string): string | null {
  let out = cleanActors(rawText)
  out = out.replace(REPR_LINE_RE, (_, key, val) =>
    `${key}: ${val.replace(/\\n$/, '').trim()}`
  )
  if (out === rawText) return null
  return out.trim()
}

function cleanGitStatus(rawText: string): string | null {
  const stripped = rawText
    .replace(STATUS_HINT_RE, '')
    .replace(/\n{3,}/g, '\n\n')
  return stripped.trim() !== rawText.trim() ? stripped.trim() : null
}

function cleanGitDiff(diffText: string, preamble: string): string | null {
  const sections = diffText.split(/(?=^diff --git )/m)
  let dropped = 0
  const kept: string[] = []

  for (const section of sections) {
    const pathMatch = section.match(/^diff --git a\/(.+?) b\//)
    const filePath  = pathMatch?.[1] ?? ''
    if (isNoisyFile(filePath)) {
      dropped++
    } else {
      kept.push(section)
    }
  }

  if (dropped === 0) return null
  kept.push(`\n[confire: ${dropped} vendored/generated file diff${dropped > 1 ? 's' : ''} omitted]`)
  return preamble + kept.join('')
}

export function optimizeGit(rawText: string): string | null {
  if (!rawText) return null

  if (rawText.startsWith('Commit history:')) return cleanGitLog(rawText)

  if (rawText.startsWith('Repository status:') || rawText.includes('On branch ')) {
    return cleanGitStatus(rawText)
  }

  if (rawText.includes('diff --git')) {
    // Header line before the first diff (e.g. "Diff with main:\n" or "Staged changes:\n")
    const headerEnd = rawText.indexOf('diff --git')
    const preamble  = rawText.slice(0, headerEnd)
    const diffBody  = rawText.slice(headerEnd)
    return cleanGitDiff(diffBody, preamble)
  }

  return null
}

export function handlesGit(event: InterceptEvent): boolean {
  return GIT_OPT_TOOLS.has(event.tool?.name ?? '')
}
