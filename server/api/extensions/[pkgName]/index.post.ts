// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '~~/shared/dto/extensions/extension.dto'
import type { ExtensionActionRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

const Body = z.object({
  action: z.enum(['install', 'uninstall', 'update']),
}) satisfies z.ZodType<ExtensionActionRequestDto>

export default defineEventHandler(async (event): Promise<ExtensionDto> => {
  const actor = event.context.authUser
  if (!actor?.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  if (!pkgName) {
    throw createError({ statusCode: 400, statusMessage: 'Missing pkgName' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  const { installExtension, uninstallExtension, updateExtension } = extensionsService()
  let extension
  if (parsed.data.action === 'install') {
    extension = await installExtension({ pkgName, actorId: actor.id })
  } else if (parsed.data.action === 'update') {
    extension = await updateExtension({ pkgName })
  } else {
    extension = await uninstallExtension({ pkgName })
  }

  return toExtensionDto(extension)
})
