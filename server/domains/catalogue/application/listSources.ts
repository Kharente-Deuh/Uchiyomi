// SPDX-License-Identifier: AGPL-3.0-or-later
import type { CatalogueRepository } from '../domain/repository'
import type { Source } from '../domain/types'

export function listSources(repo: CatalogueRepository): Promise<Source[]> {
  return repo.listSources()
}
