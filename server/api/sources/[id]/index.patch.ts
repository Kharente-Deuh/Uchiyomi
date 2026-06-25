// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateSourceRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toSourceDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

const Body = z.object({ isEnabled: z.boolean() }) satisfies z.ZodType<UpdateSourceRequestDto>

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor || !actor.canManageExtensions) {
    throw createError({ statusCode: 403, statusMessage: 'Forbidden' })
  }

  const id = getRouterParam(event, 'id')
  if (!id) {
    throw createError({ statusCode: 400, statusMessage: 'Missing id' })
  }

  const parsed = await readValidatedBody(event, Body.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid body' })
  }

  try {
    const source = await extensionsService().setSourceEnabled({ sourceId: id, isEnabled: parsed.data.isEnabled })

    return { source: toSourceDto(source) }
  } catch (err) {
    if (err instanceof Error && err.message === 'Source not found') {
      throw createError({ statusCode: 404, statusMessage: 'Source not found' })
    }

    throw err
  }
})
