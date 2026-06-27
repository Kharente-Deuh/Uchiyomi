// SPDX-License-Identifier: AGPL-3.0-or-later
import { extensionGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { toExtensionDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

export default defineEventHandler(async (event) => {
  const { extension } = await extensionGuard(event, { installationStatus: 'installed' })

  return { extension: toExtensionDto(extension) }
})
