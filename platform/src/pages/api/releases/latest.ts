import type { APIRoute } from 'astro'

interface GitHubRelease {
  tag_name: string
  name: string
  published_at: string
  html_url: string
  body: string
  assets: Array<{ name: string; browser_download_url: string }>
}

export interface ReleaseInfo {
  version: string
  publishedAt: string
  releaseUrl: string
  changelog: string
  assets: Record<string, string>
}

export const GET: APIRoute = async () => {
  const REPO = 'confire-ai/mono'
  const url = `https://api.github.com/repos/${REPO}/releases/latest`

  try {
    const res = await fetch(url, {
      headers: {
        Accept: 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'confire-platform/1.0',
      },
    })

    if (!res.ok) {
      return Response.json(
        { error: `GitHub API error: ${res.status}` },
        { status: res.status === 404 ? 404 : 502 }
      )
    }

    const release = await res.json() as GitHubRelease

    const assets: Record<string, string> = {}
    for (const asset of release.assets) {
      assets[asset.name] = asset.browser_download_url
    }

    const info: ReleaseInfo = {
      version:     release.tag_name,
      publishedAt: release.published_at,
      releaseUrl:  release.html_url,
      changelog:   release.body ?? '',
      assets,
    }

    return Response.json(info, {
      headers: {
        'Cache-Control': 'public, max-age=300, s-maxage=300',
        'Access-Control-Allow-Origin': '*',
      },
    })
  } catch (err) {
    return Response.json({ error: 'Failed to fetch release info' }, { status: 502 })
  }
}
