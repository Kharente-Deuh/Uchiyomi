// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import { z } from 'zod'
import { catalogueService } from '~~/server/domains/catalogue/application/catalogue.service'
import { toSourceSearchDto } from '~~/server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'
import { requireAuthUser } from '~~/server/domains/extensions/infrastructure/transport/http/guards/extension.guard'
import { parseQuery } from '~~/server/utils/request.util'

interface QueryParams {
  type: 'search' | 'popular' | 'latest'
  q: string
  page: number
}

// Page size is not a parameter: Suwayomi's fetchSourceManga only takes a page
// number — the source (the scraped site) dictates how many items a page holds
// via its own pagination. There is nothing to thread through here.
const QuerySchema = z.object({
  type: z.enum(['search', 'popular', 'latest']).optional().default('search'),
  q: z.string().trim().min(1).optional().default(''),
  page: z.coerce.number().int().min(1).optional().default(1),
}).superRefine((data, ctx) => {
  if (data.type === 'search' && !data.q) {
    ctx.addIssue({
      code: 'custom',
      message: 'q is required when type is search',
    })
  }
}) satisfies z.ZodType<QueryParams>

export default defineEventHandler(async (event): Promise<SourceSearchResultDto> => {
  // Overlay-only route: a visible source in the overlay already implies its
  // extension is installed, so no Suwayomi extension load is needed here — just
  // authenticate, then validate input cheaply before touching the store.
  const authUser = requireAuthUser(event)
  const { q: query, ...queryParams } = await parseQuery(event, QuerySchema)
  const pkgName = getRouterParam(event, 'pkgName')
  const sourceId = getRouterParam(event, 'id')
  if (!pkgName || !sourceId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing route params' })
  }

  // Authorize the source for this viewer. undefined → 404
  // (hides existence for unknown / wrong-pkg / disabled / NSFW-blocked alike).
  const source = await extensionsService().getVisibleSource({
    pkgName,
    sourceId,
    isAdmin: !!authUser.canManageExtensions,
    canSeeNsfw: !!authUser.allowNsfw && !!authUser.showNsfw,
  })
  if (!source) {
    throw createError({ statusCode: 404, statusMessage: 'Source not found' })
  }

  if (queryParams.type === 'latest' && !source.supportsLatest) {
    throw createError({ statusCode: 400, statusMessage: 'Source does not support latest' })
  }

  const result = await catalogueService().searchSourceWithChapters({ sourceId, query, ...queryParams })

  return toSourceSearchDto(result)
})
