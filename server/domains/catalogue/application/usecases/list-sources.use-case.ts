// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type * as Catalogue from '../../catalogue.domain'
import type * as Source from '../../source.domain'

export class UseCase implements IUseCase<void, Source.Model[]> {
  constructor(private readonly catalogueRepository: Catalogue.Repository) {}

  execute(): Promise<Source.Model[]> {
    return this.catalogueRepository.listSources()
  }
}
