// SPDX-License-Identifier: AGPL-3.0-or-later
import type { CatalogueRepository } from '../domain/repository'
import type { SearchResult } from '../domain/types'

export function searchSource(
  repo: CatalogueRepository,
  sourceId: string,
  query: string,
  page: number,
): Promise<SearchResult> {
  return repo.searchSource(sourceId, query, page)
}
