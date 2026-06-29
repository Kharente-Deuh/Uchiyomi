// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import { z } from 'zod'
import { catalogueService } from '~~/server/domains/catalogue/application/catalogue.service'
import { toSourceSearchDto } from '~~/server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { filtersRequireSearch, parseSourceFiltersParam, searchQueryMissing } from '~~/server/domains/catalogue/infrastructure/transport/http/parse-source-filters'
import { sourceGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/source.guard'
import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'
import { parseQuery } from '~~/server/utils/request.util'

// Page size is not a parameter: Suwayomi's fetchSourceManga only takes a page
// number — the source dictates how many items a page holds. `filters` is a
// JSON-encoded SourceFilterChangeDto[] (parsed and validated below); it only
// applies to the SEARCH browse type.
const QuerySchema = z.object({
  type: z.enum(['search', 'popular', 'latest']).optional().default('search'),
  q: z.string().trim().min(1).optional().default(''),
  page: z.coerce.number().int().min(1).optional().default(1),
  filters: z.string().optional(),
})

export default defineEventHandler(async (event): Promise<SourceSearchResultDto> => {
  const authUser = authGuard(event)
  const { q: query, filters: rawFilters, ...queryParams } = await parseQuery(event, QuerySchema)
  const source = await sourceGuard(event, authUser)

  if (queryParams.type === 'latest' && !source.supportsLatest) {
    throw createError({ statusCode: 400, statusMessage: 'Source does not support latest' })
  }

  const filters = parseSourceFiltersParam(rawFilters)
  if (filters === null) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid filters' })
  }

  if (filtersRequireSearch(queryParams.type, filters)) {
    throw createError({ statusCode: 400, statusMessage: 'Filters are only allowed in search mode' })
  }

  if (searchQueryMissing(queryParams.type, query, filters)) {
    throw createError({ statusCode: 400, statusMessage: 'q is required when type is search' })
  }

  const result = await catalogueService().searchSource({ sourceId: source.id, query, ...queryParams, filters })

  return toSourceSearchDto(result)
})
