// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionSettingsDto } from '~~/server/domains/extensions/infrastructure/transport/http/extension-http.presenter'
import { extensionGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'

export default defineEventHandler(async (event): Promise<ExtensionSettingsDto> => {
  const { extension } = await extensionGuard(event, {
    installationStatus: 'installed',
    mustBeAbleToManage: true,
  })

  const settings = await extensionsService().getExtensionSettings({ pkgName: extension.pkgName })

  return toExtensionSettingsDto(settings)
})
