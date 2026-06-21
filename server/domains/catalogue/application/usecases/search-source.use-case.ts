// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type * as Catalogue from '../../catalogue.domain'

export interface Opts {
  sourceId: string
  query: string
  page: number
}

export class UseCase implements IUseCase<Opts, Catalogue.SearchResult> {
  constructor(private readonly catalogueRepository: Catalogue.Repository) {}

  execute(opts: Opts): Promise<Catalogue.SearchResult> {
    return this.catalogueRepository.searchSource({ sourceId: opts.sourceId, query: opts.query, page: opts.page })
  }
}
