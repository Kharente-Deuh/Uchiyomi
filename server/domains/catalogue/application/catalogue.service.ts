// SPDX-License-Identifier: AGPL-3.0-or-later
import type { SearchMangaResult } from '../catalogue.domain'
import type { MangaDetailsModel } from '../manga.domain'
import type { SourceModel } from '../source.domain'
import type { GetMangaDetailsUseCaseParams, SearchSourceUseCaseParams } from './usecases'
import { GraphqlSuwayomiCatalogueRepository } from '../infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'
import { GetMangaDetailsUseCase, ListSourcesUseCase, SearchSourceUseCase } from './usecases'

export interface CatalogueService {
  getMangaDetails: (opts: GetMangaDetailsUseCaseParams) => Promise<MangaDetailsModel>
  listSources: () => Promise<SourceModel[]>
  searchSource: (opts: SearchSourceUseCaseParams) => Promise<SearchMangaResult>
}

const { suwayomiUrl } = useRuntimeConfig()

const suwayomiClient = createSuwayomiClient({ endpoint: `${suwayomiUrl}/api/graphql` })
const catalogueRepository = new GraphqlSuwayomiCatalogueRepository(suwayomiClient)

function getMangaDetails(opts: GetMangaDetailsUseCaseParams): Promise<MangaDetailsModel> {
  return new GetMangaDetailsUseCase(catalogueRepository).execute(opts)
}

function listSources(): Promise<SourceModel[]> {
  return new ListSourcesUseCase(catalogueRepository).execute()
}

function searchSource(opts: SearchSourceUseCaseParams): Promise<SearchMangaResult> {
  return new SearchSourceUseCase(catalogueRepository).execute(opts)
}

export function catalogueService(): CatalogueService {
  return {
    getMangaDetails,
    listSources,
    searchSource,
  }
}
