// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { CatalogueRepository, SearchMangaResult, SourceBrowseType } from '../../catalogue.domain'

export interface SearchSourceUseCaseParams {
  sourceId: string
  query: string
  page: number
  type: SourceBrowseType
}

export class SearchSourceUseCase implements IUseCase<SearchSourceUseCaseParams, SearchMangaResult> {
  constructor(private readonly catalogueRepository: CatalogueRepository) {}

  execute(opts: SearchSourceUseCaseParams): Promise<SearchMangaResult> {
    return this.catalogueRepository.searchSource({ sourceId: opts.sourceId, query: opts.query, page: opts.page, type: opts.type })
  }
}
