// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SearchMangaResult, SearchSourceWithChaptersResult } from '../catalogue.domain'
import type { MangaDetailsModel } from '../manga.domain'
import type { SourceModel } from '../source.domain'
import type { GetMangaDetailsUseCaseParams, SearchSourceUseCaseParams, SearchSourceWithChaptersUseCaseParams } from './usecases'
import { GraphqlSuwayomiCatalogueRepository } from '../infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'
import { GetMangaDetailsUseCase, ListSourcesUseCase, SearchSourceUseCase, SearchSourceWithChaptersUseCase } from './usecases'

export interface CatalogueService {
  getMangaDetails: (opts: GetMangaDetailsUseCaseParams) => Promise<MangaDetailsModel>
  listSources: () => Promise<SourceModel[]>
  searchSource: (opts: SearchSourceUseCaseParams) => Promise<SearchMangaResult>
  searchSourceWithChapters: (opts: SearchSourceWithChaptersUseCaseParams) => Promise<SearchSourceWithChaptersResult>
}

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const catalogueRepository = new GraphqlSuwayomiCatalogueRepository(suwayomiClient)

// Bounded in-flight chapter fetches per search page (each hits the remote source).
const SEARCH_ENRICH_CONCURRENCY = 5

function getMangaDetails(opts: GetMangaDetailsUseCaseParams): Promise<MangaDetailsModel> {
  return new GetMangaDetailsUseCase(catalogueRepository).execute(opts)
}

function listSources(): Promise<SourceModel[]> {
  return new ListSourcesUseCase(catalogueRepository).execute()
}

function searchSource(opts: SearchSourceUseCaseParams): Promise<SearchMangaResult> {
  return new SearchSourceUseCase(catalogueRepository).execute(opts)
}

function searchSourceWithChapters(opts: SearchSourceWithChaptersUseCaseParams): Promise<SearchSourceWithChaptersResult> {
  return new SearchSourceWithChaptersUseCase(catalogueRepository, SEARCH_ENRICH_CONCURRENCY).execute(opts)
}

export function catalogueService(): CatalogueService {
  return {
    getMangaDetails,
    listSources,
    searchSource,
    searchSourceWithChapters,
  }
}
