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

  const { getExtensionHealth, getExtensionByPkgName } = extensionsService()
  const extension = await getExtensionByPkgName({ pkgName })
  if (!extension) {
    throw createError({ statusCode: 404, statusMessage: 'Extension not found' })
  }

  if (!extension.isInstalled || extension.hasUpdate) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  if (!actor.allowNsfw && extension.isNsfw) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const healthResult = await getExtensionHealth({ pkgName })

  return {
    extension: toExtensionDto(extension),
    health: healthResult ? toHealthDto(healthResult.health, healthResult.log) : null,
  }
})
