interface Env {
  RELEASES: R2Bucket
}

const GET_HOST = 'get.confire.dev'
const RELEASES_HOST = 'releases.confire.dev'

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      return new Response('Method not allowed', { status: 405 })
    }

    if (url.hostname === GET_HOST) {
      return handleGetHost(url.pathname, env, request.method)
    }

    if (url.hostname === RELEASES_HOST) {
      return handleReleasesHost(url.pathname, env, request.method)
    }

    return new Response('Not found', { status: 404 })
  },
} satisfies ExportedHandler<Env>

async function handleGetHost(pathname: string, env: Env, method: string): Promise<Response> {
  if (pathname === '/' || pathname === '/install.sh') {
    return serveObject(env, 'install.sh', 'text/x-shellscript; charset=utf-8', 'public, max-age=300', method)
  }
  if (pathname === '/latest.json') {
    return serveObject(env, 'latest.json', 'application/json; charset=utf-8', 'public, max-age=60', method)
  }
  return new Response('Not found', { status: 404 })
}

async function handleReleasesHost(pathname: string, env: Env, method: string): Promise<Response> {
  const key = pathname.replace(/^\//, '')
  if (!key || key.includes('..') || key.includes('//')) {
    return new Response('Not found', { status: 404 })
  }

  const cache = key === 'latest.json'
    ? 'public, max-age=60'
    : 'public, max-age=31536000, immutable'

  return serveObject(env, key, contentType(key), cache, method)
}

async function serveObject(
  env: Env,
  key: string,
  contentType: string,
  cacheControl: string,
  method: string,
): Promise<Response> {
  const object = await env.RELEASES.get(key)
  if (!object) {
    return new Response('Not found', { status: 404 })
  }

  const headers = new Headers()
  object.writeHttpMetadata(headers)
  headers.set('Content-Type', contentType)
  headers.set('Cache-Control', cacheControl)
  headers.set('Access-Control-Allow-Origin', '*')
  if (object.httpEtag) {
    headers.set('ETag', object.httpEtag)
  }

  if (method === 'HEAD') {
    return new Response(null, { status: 200, headers })
  }

  return new Response(object.body, { status: 200, headers })
}

function contentType(key: string): string {
  if (key.endsWith('.json')) return 'application/json; charset=utf-8'
  if (key.endsWith('.txt')) return 'text/plain; charset=utf-8'
  if (key.endsWith('.sh')) return 'text/x-shellscript; charset=utf-8'
  return 'application/octet-stream'
}
