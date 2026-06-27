// SPDX-License-Identifier: AGPL-3.0-or-later
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toSourceDto } from '~~/server/domains/extensions/infrastructure/transport/http/extension-http.presenter'
import { requireAuthUser } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'

export default defineEventHandler(async (event) => {
  // Overlay-only route: sources are mirrored per installed extension, so listing
  // them needs no Suwayomi extension load — authenticate and query the overlay,
  // which returns an empty list for an unknown / not-installed pkgName.
  const authUser = requireAuthUser(event)
  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const sources = await extensionsService().listExtensionSources({
    pkgName,
    isAdmin: !!authUser.canManageExtensions,
    canSeeNsfw: !!authUser.allowNsfw && !!authUser.showNsfw,
  })

  return { sources: sources.map(s => toSourceDto(s)) }
})
