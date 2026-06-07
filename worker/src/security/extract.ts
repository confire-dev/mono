// extractText pulls the plain text content from any tool output shape.
// Handles MCP envelope {content:[{type:'text',text:...}]} and plain strings.
export function extractText(output: unknown): string | null {
  if (!output) return null
  if (typeof output === 'string') return output || null
  if (typeof output !== 'object') return null
  const o = output as Record<string, unknown>
  // MCP envelope: {content:[{type:'text',text:'...'}]}
  if (Array.isArray(o['content'])) {
    const texts: string[] = []
    for (const item of o['content'] as unknown[]) {
      if (item && typeof item === 'object') {
        const c = item as Record<string, unknown>
        if (c['type'] === 'text' && typeof c['text'] === 'string') {
          texts.push(c['text'])
        }
      }
    }
    return texts.join('\n') || null
  }
  // Bash stdout envelope: {stdout:'...', stderr:'...'}
  if (typeof o['stdout'] === 'string') return o['stdout'] || null
  return null
}
