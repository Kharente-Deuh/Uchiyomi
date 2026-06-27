// SPDX-License-Identifier: AGPL-3.0-or-later
import { Buffer } from 'node:buffer'

// Proxies a manga cover from the internal Suwayomi server (ADR-0001: Suwayomi is
// never exposed to clients). Suwayomi serves cover bytes at
// `/api/v1/manga/<id>/thumbnail`; we stream them back behind auth. Covers can change
// over time (unlike per-version extension icons), so the cache is time-bounded, not
// immutable. No per-series NSFW gate — NSFW is a source-level concern and is enforced
// at search time.
export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const id = getRouterParam(event, 'id')
  if (!id || !/^\d+$/.test(id)) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid manga id' })
  }

  const { suwayomiUrl } = useRuntimeConfig(event)
  const upstream = await fetch(`${suwayomiUrl}/api/v1/manga/${id}/thumbnail`)
  if (!upstream.ok) {
    throw createError({ statusCode: 502, statusMessage: 'Failed to fetch thumbnail from Suwayomi' })
  }

  setResponseHeader(event, 'Content-Type', upstream.headers.get('content-type') ?? 'image/jpeg')
  setResponseHeader(event, 'Cache-Control', 'public, max-age=86400')

  return Buffer.from(await upstream.arrayBuffer())
})
