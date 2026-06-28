// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '../../../../shared/use-case'
import type { CatalogueRepository, SearchedSourceMangaItem, SearchSourceWithChaptersResult, SourceBrowseType } from '../../catalogue.domain'

export interface SearchSourceWithChaptersUseCaseParams {
  sourceId: string
  query: string
  page: number
  type: SourceBrowseType
}

// Run an async mapper over items with a bounded number of in-flight calls,
// preserving input order. Keeps the N+1 chapter enrichment from hammering the
// source / the engine.
async function mapWithConcurrency<T, R>(items: T[], limit: number, fn: (item: T) => Promise<R>): Promise<R[]> {
  const results = Array.from<R>({ length: items.length })
  let next = 0

  async function worker(): Promise<void> {
    while (next < items.length) {
      const i = next++
      results[i] = await fn(items[i]!)
    }
  }

  const workers = Array.from({ length: Math.min(limit, items.length) }, () => worker())
  await Promise.all(workers)

  return results
}

export class SearchSourceWithChaptersUseCase implements IUseCase<SearchSourceWithChaptersUseCaseParams, SearchSourceWithChaptersResult> {
  constructor(
    private readonly catalogueRepository: CatalogueRepository,
    private readonly concurrency = 5,
  ) {}

  async execute(opts: SearchSourceWithChaptersUseCaseParams): Promise<SearchSourceWithChaptersResult> {
    const { mangas, hasNextPage } = await this.catalogueRepository.searchSource(opts)

    const items = await mapWithConcurrency<typeof mangas[number], SearchedSourceMangaItem>(
      mangas,
      this.concurrency,
      async (m) => {
        try {
          const summary = await this.catalogueRepository.getSourceMangaChapterSummary(m.id)

          return {
            id: m.id,
            title: m.title,
            thumbnailUrl: m.thumbnailUrl,
            inLibrary: m.inLibrary,
            chapterCount: summary.chapterCount,
            lastChapter: summary.lastChapter,
          }
        } catch {
          // Per-item resilience: a failed enrichment degrades to null fields,
          // it must never abort the whole page.
          return {
            id: m.id,
            title: m.title,
            thumbnailUrl: m.thumbnailUrl,
            inLibrary: m.inLibrary,
            chapterCount: null,
            lastChapter: null,
          }
        }
      },
    )

    return { items, hasNextPage }
  }
}
