// SPDX-License-Identifier: AGPL-3.0-or-later
import { Buffer } from 'node:buffer'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'

// Proxies an extension icon from the internal Suwayomi server (ADR-0001: Suwayomi
// is never exposed to clients). Suwayomi's `/api/v1/extension/icon/<apk>` route
// returns image bytes despite the `.apk` path; we stream them back behind auth
// with a long-lived cache header (icons are immutable per extension version).
export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const extension = await extensionsService().getExtensionByPkgName({ pkgName })
  if (!extension) {
    throw createError({ statusCode: 404, statusMessage: 'Extension not found' })
  }

  if (!actor.canManageExtensions && !extension.isInstalled) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  if (!actor.allowNsfw && extension.isNsfw) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const { suwayomiUrl } = useRuntimeConfig(event)
  const upstream = await fetch(`${suwayomiUrl}${extension.iconUrl}`)
  if (!upstream.ok) {
    throw createError({ statusCode: 502, statusMessage: 'Failed to fetch icon from Suwayomi' })
  }

  setResponseHeader(event, 'Content-Type', upstream.headers.get('content-type') ?? 'image/png')
  setResponseHeader(event, 'Cache-Control', 'public, max-age=31536000, immutable')

  return Buffer.from(await upstream.arrayBuffer())
})
