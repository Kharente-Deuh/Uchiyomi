// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionActionRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { installExtension, uninstallExtension } from '~~/server/domains/extensions/application'

const Body = z.object({ action: z.enum(['install', 'uninstall']) }) satisfies z.ZodType<ExtensionActionRequestDto>

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor || !actor.canManageExtensions) {
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

  if (parsed.data.action === 'install') {
    await installExtension.execute({ pkgName, actorId: actor.id })
  } else {
    await uninstallExtension.execute({ pkgName })
  }

  return { ok: true }
})
