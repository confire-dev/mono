interface Env {
  RELEASES: R2Bucket
}

const GET_HOST     = 'get.confire.dev'
const GET_DEV_HOST = 'get.dev.confire.dev'
const RELEASES_HOST = 'releases.confire.dev'
const LOCAL_HOSTS = new Set(['localhost', '127.0.0.1', '[::1]'])

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      return new Response('Method not allowed', { status: 405 })
    }

    if (LOCAL_HOSTS.has(url.hostname)) {
      return handleLocalDev(url.pathname, env, request.method)
    }

    if (url.hostname === GET_HOST) {
      return handleGetHost(url.pathname, env, request.method)
    }

    if (url.hostname === GET_DEV_HOST) {
      return handleGetDevHost(url.pathname, env, request.method)
    }

    if (url.hostname === RELEASES_HOST) {
      return handleReleasesHost(url.pathname, env, request.method)
    }

    return new Response('Not found', { status: 404 })
  },
} satisfies ExportedHandler<Env>

/** Single-port local dev: /latest.json + /install.sh vs /{version}/binary paths. */
async function handleLocalDev(pathname: string, env: Env, method: string): Promise<Response> {
  if (pathname === '/' || pathname === '/install.sh') {
    return serveObject(env, 'install.sh', 'text/x-shellscript; charset=utf-8', 'public, max-age=300', method)
  }
  if (pathname === '/latest.json') {
    return serveObject(env, 'latest.json', 'application/json; charset=utf-8', 'public, max-age=60', method)
  }
  const key = pathname.replace(/^\//, '')
  if (!isSafeObjectKey(key)) {
    return new Response('Not found', { status: 404 })
  }
  const cache = key.endsWith('latest.json')
    ? 'public, max-age=60'
    : 'public, max-age=31536000, immutable'
  return serveObject(env, key, contentType(key), cache, method)
}

async function handleGetHost(pathname: string, env: Env, method: string): Promise<Response> {
  if (pathname === '/' || pathname === '/install.sh') {
    return serveObject(env, 'install.sh', 'text/x-shellscript; charset=utf-8', 'public, max-age=300', method)
  }
  if (pathname === '/latest.json') {
    return serveObject(env, 'latest.json', 'application/json; charset=utf-8', 'public, max-age=60', method)
  }
  return new Response('Not found', { status: 404 })
}

function isSafeObjectKey(key: string): boolean {
  if (!key || key.startsWith('/') || key.includes('//')) return false
  for (const segment of key.split('/')) {
    if (segment === '..' || segment === '.') return false
  }
  return true
}

// Dev channel: get.dev.confire.dev — artifacts stored under dev/ prefix in R2.
// install.sh and latest.json live at dev/install.sh, dev/latest.json.
// Binaries: dev/releases/<version>/confire_<version>_<os>_<arch>.tar.gz
async function handleGetDevHost(pathname: string, env: Env, method: string): Promise<Response> {
  if (pathname === '/' || pathname === '/install.sh') {
    return serveObject(env, 'dev/install.sh', 'text/x-shellscript; charset=utf-8', 'public, max-age=300', method)
  }
  if (pathname === '/latest.json') {
    return serveObject(env, 'dev/latest.json', 'application/json; charset=utf-8', 'no-cache', method)
  }
  const key = `dev${pathname}`
  if (!isSafeObjectKey(key)) {
    return new Response('Not found', { status: 404 })
  }
  return serveObject(env, key, contentType(key), 'public, max-age=31536000, immutable', method)
}

async function handleReleasesHost(pathname: string, env: Env, method: string): Promise<Response> {
  const key = pathname.replace(/^\//, '')
  if (!isSafeObjectKey(key)) {
    return new Response('Not found', { status: 404 })
  }

  const cache = key.endsWith('latest.json')
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
