// SPDX-License-Identifier: AGPL-3.0-or-later
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionDto, toHealthDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const { getExtensionHealth, listExtensions } = extensionsService()
  const { items } = await listExtensions({
    isAdmin: !!actor.canManageExtensions,
    viewerCanSeeNsfw: !!actor.allowNsfw && !!actor.showNsfw,
    page: 1,
    pageSize: 1,
    filters: { pkgName },
  })
  const extension = items[0]
  if (!extension) {
    throw createError({ statusCode: 404, statusMessage: 'Extension not available' })
  }

  const healthResult = await getExtensionHealth({ pkgName })

  return {
    extension: toExtensionDto(extension),
    health: healthResult ? toHealthDto(healthResult.health, healthResult.log) : null,
  }
})
