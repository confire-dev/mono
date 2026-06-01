import type { APIRoute } from 'astro'

export interface ReleaseInfo {
  version: string
  publishedAt: string
  releaseUrl: string
  changelog: string
  assets: Record<string, string>
}

const MANIFEST_URL = 'https://get.confire.dev/latest.json'

export const GET: APIRoute = async () => {
  try {
    const res = await fetch(MANIFEST_URL, {
      headers: { Accept: 'application/json' },
    })

    if (!res.ok) {
      return Response.json(
        { error: `Release manifest error: ${res.status}` },
        { status: res.status === 404 ? 404 : 502 },
      )
    }

    const manifest = await res.json() as {
      version: string
      publishedAt?: string
      releaseUrl?: string
      assets?: Record<string, string>
    }

    const info: ReleaseInfo = {
      version: manifest.version,
      publishedAt: manifest.publishedAt ?? '',
      releaseUrl: manifest.releaseUrl ?? `https://releases.confire.dev/${manifest.version}`,
      changelog: '',
      assets: manifest.assets ?? {},
    }

    return Response.json(info, {
      headers: {
        'Cache-Control': 'public, max-age=300, s-maxage=300',
        'Access-Control-Allow-Origin': '*',
      },
    })
  } catch {
    return Response.json({ error: 'Failed to fetch release info' }, { status: 502 })
  }
}
