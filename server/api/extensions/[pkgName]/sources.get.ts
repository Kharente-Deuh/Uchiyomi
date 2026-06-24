// SPDX-License-Identifier: AGPL-3.0-or-later
import { listExtensionSources } from '~~/server/domains/extensions/application'
import { toSourceDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const sources = await listExtensionSources.execute({
    pkgName,
    isAdmin: !!actor.canManageExtensions,
    viewerCanSeeNsfw: !!actor.allowNsfw && !!actor.showNsfw,
  })

  return { sources: sources.map(s => toSourceDto(s)) }
})
