// SPDX-License-Identifier: AGPL-3.0-or-later
import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { toExtensionListResponseDto } from '../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

// `?flag=true|false` arrives as a string; coerce to an optional boolean.
const BoolParam = z.enum(['true', 'false']).transform(v => v === 'true').optional()

const Query = z.object({
  page: z.coerce.number().int().min(1).default(1),
  pageSize: z.coerce.number().int().min(1).max(100).default(20),
  search: z.string().trim().min(1).optional(),
  isInstalled: BoolParam,
  hasUpdate: BoolParam,
  nsfw: BoolParam,
})

export default defineEventHandler(async (event) => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const parsed = await getValidatedQuery(event, Query.safeParse)
  if (!parsed.success) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid query' })
  }

  const { page, pageSize, search, isInstalled, hasUpdate, nsfw } = parsed.data

  const result = await extensionsService().listExtensions({
    isAdmin: !!actor.canManageExtensions,
    viewerCanSeeNsfw: !!actor.allowNsfw && !!actor.showNsfw,
    page,
    pageSize,
    filters: { search, isInstalled, hasUpdate, nsfw },
  })

  return toExtensionListResponseDto(result)
})
