// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { CatalogueRepository } from '../../catalogue.domain'
import type { MangaDetailsModel } from '../../manga.domain'

export interface GetMangaDetailsUseCaseParams {
  mangaId: string
}

export class GetMangaDetailsUseCase implements IUseCase<GetMangaDetailsUseCaseParams, MangaDetailsModel> {
  constructor(private readonly catalogueRepository: CatalogueRepository) {}

  execute(opts: GetMangaDetailsUseCaseParams): Promise<MangaDetailsModel> {
    return this.catalogueRepository.getMangaDetails({ mangaId: opts.mangaId })
  }
}
