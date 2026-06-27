// SPDX-License-Identifier: AGPL-3.0-or-later
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { extensionGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { toExtensionDto, toHealthDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event) => {
  const { extension } = await extensionGuard(event, { installationStatus: 'installed' })
  const healthResult = await extensionsService().getExtensionHealth({ pkgName: extension.pkgName })

  return {
    extension: toExtensionDto(extension),
    health: healthResult ? toHealthDto(healthResult.health, healthResult.log) : null,
  }
})
