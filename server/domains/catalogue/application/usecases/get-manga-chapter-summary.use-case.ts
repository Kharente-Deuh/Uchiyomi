// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { CatalogueRepository } from '../../catalogue.domain'
import type { MangaChapterSummaryModel } from '../../manga.domain'

export interface GetMangaChapterSummaryParams {
  mangaId: string
}

export class GetMangaChapterSummaryUseCase implements IUseCase<GetMangaChapterSummaryParams, MangaChapterSummaryModel | undefined> {
  constructor(private readonly catalogueRepository: CatalogueRepository) {}

  execute({ mangaId }: GetMangaChapterSummaryParams): Promise<MangaChapterSummaryModel | undefined> {
    return this.catalogueRepository.getMangaChapterSummary(mangaId)
  }
}
