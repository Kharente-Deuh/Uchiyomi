// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SearchMangaResult, SourceFilter } from '../catalogue.domain'
import type { MangaChapterSummaryModel, MangaDetailsModel } from '../manga.domain'
import type { SourceModel } from '../source.domain'
import type { GetMangaDetailsUseCaseParams, GetSourceFiltersUseCaseParams, SearchSourceUseCaseParams } from './usecases'
import { GraphqlSuwayomiCatalogueRepository } from '../infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'
import { GetMangaChapterSummaryUseCase, GetMangaDetailsUseCase, GetSourceFiltersUseCase, ListSourcesUseCase, SearchSourceUseCase } from './usecases'

export interface CatalogueService {
  getMangaDetails: (opts: GetMangaDetailsUseCaseParams) => Promise<MangaDetailsModel>
  listSources: () => Promise<SourceModel[]>
  searchSource: (opts: SearchSourceUseCaseParams) => Promise<SearchMangaResult>
  getMangaChapterSummary: (mangaId: string) => Promise<MangaChapterSummaryModel | undefined>
  getSourceFilters: (opts: GetSourceFiltersUseCaseParams) => Promise<SourceFilter[]>
}

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const catalogueRepository = new GraphqlSuwayomiCatalogueRepository(suwayomiClient)

function getMangaDetails(opts: GetMangaDetailsUseCaseParams): Promise<MangaDetailsModel> {
  return new GetMangaDetailsUseCase(catalogueRepository).execute(opts)
}

function getMangaChapterSummary(mangaId: string): Promise<MangaChapterSummaryModel | undefined> {
  return new GetMangaChapterSummaryUseCase(catalogueRepository).execute({ mangaId })
}

function listSources(): Promise<SourceModel[]> {
  return new ListSourcesUseCase(catalogueRepository).execute()
}

function searchSource(opts: SearchSourceUseCaseParams): Promise<SearchMangaResult> {
  return new SearchSourceUseCase(catalogueRepository).execute(opts)
}

function getSourceFilters(opts: GetSourceFiltersUseCaseParams): Promise<SourceFilter[]> {
  return new GetSourceFiltersUseCase(catalogueRepository).execute(opts)
}

export function catalogueService(): CatalogueService {
  return {
    getMangaDetails,
    listSources,
    searchSource,
    getMangaChapterSummary,
    getSourceFilters,
  }
}
