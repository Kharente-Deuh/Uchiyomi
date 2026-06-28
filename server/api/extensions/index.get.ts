// SPDX-License-Identifier: AGPL-3.0-or-later

import { z } from 'zod'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { requireAuthUser } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { toPageDto } from '~~/server/shared'
import { parseQuery } from '~~/server/utils/request.util'
import { toExtensionDto } from '../../domains/extensions/infrastructure/transport/http/extension-http.presenter'

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
  const authUser = requireAuthUser(event)
  const { page, pageSize, ...filters } = await parseQuery(event, Query)

  const result = await extensionsService().listExtensions({
    isAdmin: !!authUser.canManageExtensions,
    viewerCanSeeNsfw: !!authUser.allowNsfw && !!authUser.showNsfw,
    page,
    pageSize,
    filters,
  })

  return toPageDto(result, toExtensionDto)
})
