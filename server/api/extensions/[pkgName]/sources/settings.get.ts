// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionSettingsDto } from '~~/server/domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event): Promise<ExtensionSettingsDto> => {
  const actor = event.context.authUser
  if (!actor?.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const settings = await extensionsService().getExtensionSettings({ pkgName })

  return toExtensionSettingsDto(settings)
})
