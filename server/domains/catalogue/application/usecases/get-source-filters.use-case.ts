// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { CatalogueRepository, SourceFilter } from '../../catalogue.domain'

export interface GetSourceFiltersUseCaseParams {
  sourceId: string
}

export class GetSourceFiltersUseCase implements IUseCase<GetSourceFiltersUseCaseParams, SourceFilter[]> {
  constructor(private readonly catalogueRepository: CatalogueRepository) {}

  execute(opts: GetSourceFiltersUseCaseParams): Promise<SourceFilter[]> {
    return this.catalogueRepository.getSourceFilters(opts.sourceId)
  }
}
