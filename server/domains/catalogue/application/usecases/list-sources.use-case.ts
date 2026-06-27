// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { CatalogueRepository } from '../../catalogue.domain'
import type { SourceModel } from '../../source.domain'

export class ListSourcesUseCase implements IUseCase<void, SourceModel[]> {
  constructor(private readonly catalogueRepository: CatalogueRepository) {}

  execute(): Promise<SourceModel[]> {
    return this.catalogueRepository.listSources()
  }
}
