// SPDX-License-Identifier: AGPL-3.0-or-later
import { Buffer } from 'node:buffer'
import { extensionGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'

// Proxies an extension icon from the internal Suwayomi server (ADR-0001: Suwayomi
// is never exposed to clients). Suwayomi's `/api/v1/extension/icon/<apk>` route
// returns image bytes despite the `.apk` path; we stream them back behind auth
// with a long-lived cache header (icons are immutable per extension version).
export default defineEventHandler(async (event) => {
  // Managers may preview the icon of any extension (incl. not-yet-installed, to
  // browse the store); regular users only see icons of installed extensions.
  // Icons are version-immutable, so an available update is irrelevant here.
  const { extension } = await extensionGuard(event, {
    installationStatus: 'installed',
    byPassUpdateCheck: true,
    adminBypassesInstallCheck: true,
  })
  const { suwayomiUrl } = useRuntimeConfig(event)
  const upstream = await fetch(`${suwayomiUrl}${extension.iconUrl}`)
  if (!upstream.ok) {
    throw createError({ statusCode: 502, statusMessage: 'Failed to fetch icon from Suwayomi' })
  }

  setResponseHeader(event, 'Content-Type', upstream.headers.get('content-type') ?? 'image/png')
  setResponseHeader(event, 'Cache-Control', 'public, max-age=31536000, immutable')

  return Buffer.from(await upstream.arrayBuffer())
})
