// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SourceSearchResultDto } from '#shared/dto/catalogue/source-search.dto'
import { catalogueService } from '~~/server/domains/catalogue/application/catalogue.service'
import { toSourceSearchDto } from '~~/server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { extensionsService } from '~~/server/domains/extensions/application/extensions.service'

export default defineEventHandler(async (event): Promise<SourceSearchResultDto> => {
  const actor = event.context.authUser
  if (!actor) {
    throw createError({ statusCode: 401, statusMessage: 'Unauthenticated' })
  }

  const pkgName = getRouterParam(event, 'pkgName')
  const sourceId = getRouterParam(event, 'id')
  if (!pkgName || !sourceId) {
    throw createError({ statusCode: 400, statusMessage: 'Missing route params' })
  }

  // Validate query params before the DB/Suwayomi call so bad input is rejected cheaply.
  const queryParams = getQuery(event)
  const typeParam = queryParams.type === undefined ? 'search' : queryParams.type
  if (typeParam !== 'search' && typeParam !== 'popular' && typeParam !== 'latest') {
    throw createError({ statusCode: 400, statusMessage: 'Invalid type' })
  }

  const q = typeof queryParams.q === 'string' ? queryParams.q : ''
  const page = queryParams.page === undefined ? 1 : Number(queryParams.page)
  if (!Number.isInteger(page) || page < 1) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid page' })
  }

  // Authorize the source for this viewer. undefined → 404
  // (hides existence for unknown / wrong-pkg / disabled / NSFW-blocked alike).
  const source = await extensionsService().getVisibleSource({
    pkgName,
    sourceId,
    isAdmin: !!actor.canManageExtensions,
    canSeeNsfw: !!actor.allowNsfw && !!actor.showNsfw,
  })
  if (!source) {
    throw createError({ statusCode: 404, statusMessage: 'Source not found' })
  }

  if (typeParam === 'latest' && !source.supportsLatest) {
    throw createError({ statusCode: 400, statusMessage: 'Source does not support latest' })
  }

  const result = await catalogueService().searchSourceWithChapters({ sourceId, query: q, page, type: typeParam })

  return toSourceSearchDto(result)
})
