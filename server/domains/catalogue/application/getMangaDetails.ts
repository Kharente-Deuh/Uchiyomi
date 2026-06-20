// SPDX-License-Identifier: AGPL-3.0-or-later
import type { CatalogueRepository } from '../domain/repository'
import type { MangaDetails } from '../domain/types'

export function getMangaDetails(repo: CatalogueRepository, mangaId: string): Promise<MangaDetails> {
  return repo.getMangaDetails(mangaId)
}
