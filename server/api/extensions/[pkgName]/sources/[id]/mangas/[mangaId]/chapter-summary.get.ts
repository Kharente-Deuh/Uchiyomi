// SPDX-License-Identifier: AGPL-3.0-or-later

import type { MangaChapterSummaryDto } from '~~/shared/dto/catalogue/manga-chapter-summary.dto'
import { catalogueService } from '~~/server/domains/catalogue/application/catalogue.service'
import { toMangaChapterSummaryDto } from '~~/server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'
import { sourceGuard } from '~~/server/domains/extensions/infrastructure/transport/http/guards/source.guard'
import { authGuard } from '~~/server/domains/identity/auth/infrastructure/http/guards/auth.guard'

export default defineEventHandler(async (event): Promise<MangaChapterSummaryDto> => {
  const authUser = authGuard(event)
  const mangaId = getRouterParam(event, 'mangaId')
  if (!mangaId || Number.isNaN(Number(mangaId))) {
    throw createError({ statusCode: 400, statusMessage: 'Invalid mangaId' })
  }

  await sourceGuard(event, authUser)

  const summary = await catalogueService().getMangaChapterSummary(mangaId)
  if (!summary) {
    throw createError({ statusCode: 404, statusMessage: 'Manga chapter summary not found' })
  }

  return toMangaChapterSummaryDto(summary)
})
