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

  const iconPath = await extensionsService().resolveExtensionIconUrl(pkgName, {
    isAdmin: !!actor.canManageExtensions,
    viewerCanSeeNsfw: !!actor.allowNsfw && !!actor.showNsfw,
  })
  if (!iconPath) {
    throw createError({ statusCode: 404, statusMessage: 'Icon not found' })
  }

  const { suwayomiUrl } = useRuntimeConfig(event)
  const upstream = await fetch(`${suwayomiUrl}${iconPath}`)
  if (!upstream.ok) {
    throw createError({ statusCode: 502, statusMessage: 'Failed to fetch icon from Suwayomi' })
  }

  setResponseHeader(event, 'Content-Type', upstream.headers.get('content-type') ?? 'image/png')
  setResponseHeader(event, 'Cache-Control', 'public, max-age=31536000, immutable')

  return Buffer.from(await upstream.arrayBuffer())
})
