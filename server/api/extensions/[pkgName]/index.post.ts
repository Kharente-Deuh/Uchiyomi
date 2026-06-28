// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionModel } from '~~/server/domains/extensions/extension.domain'
import type { ExtensionDto } from '~~/shared/dto/extensions/extension.dto'
import type { ExtensionActionRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { requireAuthUser, requireExtension } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { parseBody } from '~~/server/utils/request.util'
import { toExtensionDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

const BodySchema = z.object({
  action: z.enum(['install', 'uninstall', 'update']),
}) satisfies z.ZodType<ExtensionActionRequestDto>

export default defineEventHandler(async (event): Promise<ExtensionDto> => {
  const authUser = requireAuthUser(event, { mustBeAbleToManage: true })
  const body = await parseBody(event, BodySchema)
  const extension = await requireExtension(event, authUser, {
    installationStatus: body.action === 'install' ? 'not-installed' : 'installed',
    byPassUpdateCheck: body.action === 'update',
  })

  const { installExtension, uninstallExtension, updateExtension } = extensionsService()
  let updatedExtension: ExtensionModel | undefined

  switch (body.action) {
    case 'install':
      updatedExtension = await installExtension({ pkgName: extension.pkgName, actorId: authUser.id })
      break
    case 'update':
      updatedExtension = await updateExtension({ pkgName: extension.pkgName })
      break
    case 'uninstall':
      updatedExtension = await uninstallExtension({ pkgName: extension.pkgName })
      break
  }

  return toExtensionDto(updatedExtension)
})
