// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type * as Catalogue from '../../catalogue.domain'
import type * as Manga from '../../manga.domain'

export interface Opts {
  mangaId: string
}

export class UseCase implements IUseCase<Opts, Manga.Details> {
  constructor(private readonly catalogueRepository: Catalogue.Repository) {}

  execute(opts: Opts): Promise<Manga.Details> {
    return this.catalogueRepository.getMangaDetails({ mangaId: opts.mangaId })
  }
}
