// SPDX-License-Identifier: AGPL-3.0-or-later
import type { UpdateSourcePreferenceRequestDto } from '#shared/dto/extensions/extensions.request'
import { z } from 'zod'
import { updateSourcePreference } from '~~/server/domains/extensions/application'
import { toPreferenceDto } from '../../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

const Body = z.object({
  position: z.number().int().nonnegative(),
  booleanValue: z.boolean().optional(),
  textValue: z.string().optional(),
  multiValue: z.array(z.string()).optional(),
}) satisfies z.ZodType<UpdateSourcePreferenceRequestDto>

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

  const prefs = await updateSourcePreference.execute({
    sourceId: id,
    position: parsed.data.position,
    booleanValue: parsed.data.booleanValue,
    textValue: parsed.data.textValue,
    multiValue: parsed.data.multiValue,
  })

  return { preferences: prefs.map(p => toPreferenceDto(p)) }
})
