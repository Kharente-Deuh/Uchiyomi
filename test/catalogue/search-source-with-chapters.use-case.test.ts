// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import { SearchSourceWithChaptersUseCase } from '../../server/domains/catalogue/application/usecases/search-source-with-chapters.use-case'

function makeRepo(overrides: Partial<Record<'searchSource' | 'getSourceMangaChapterSummary', ReturnType<typeof vi.fn>>> = {}): never {
  const searchSource = overrides.searchSource ?? vi.fn().mockResolvedValue({
    mangas: [
      { id: '1', title: 'A', thumbnailUrl: 'a.jpg', inLibrary: false },
      { id: '2', title: 'B', thumbnailUrl: undefined, inLibrary: true },
    ],
    hasNextPage: true,
  })
  const getSourceMangaChapterSummary = overrides.getSourceMangaChapterSummary ?? vi.fn(async (id: string) => ({
    chapterCount: id === '1' ? 10 : 20,
    lastChapter: { name: `last-${id}`, uploadDate: '1000' },
  }))

  return { searchSource, getSourceMangaChapterSummary } as never
}

describe('searchSourceWithChaptersUseCase', () => {
  it('delegates the search with the given params and passes hasNextPage through', async () => {
    const repo = makeRepo()
    const uc = new SearchSourceWithChaptersUseCase(repo)
    const res = await uc.execute({ sourceId: 's1', query: 'naruto', page: 2, type: 'search' })

    expect(repo.searchSource).toHaveBeenCalledWith({ sourceId: 's1', query: 'naruto', page: 2, type: 'search' })
    expect(res.hasNextPage).toBe(true)
    expect(res.items.map(i => [i.id, i.chapterCount, i.lastChapter?.name])).toEqual([
      ['1', 10, 'last-1'],
      ['2', 20, 'last-2'],
    ])
  })

  it('enriches every result item, preserving order', async () => {
    const repo = makeRepo()
    const uc = new SearchSourceWithChaptersUseCase(repo, 1)
    const res = await uc.execute({ sourceId: 's1', query: '', page: 1, type: 'search' })

    expect(repo.getSourceMangaChapterSummary).toHaveBeenCalledTimes(2)
    expect(res.items[0]!.thumbnailUrl).toBe('a.jpg')
    expect(res.items[1]!.thumbnailUrl).toBeUndefined()
  })

  it('degrades a failed enrichment to null fields without aborting the page', async () => {
    const getSourceMangaChapterSummary = vi.fn(async (id: string) => {
      if (id === '1') {
        throw new Error('source unreachable')
      }

      return { chapterCount: 20, lastChapter: { name: 'last-2', uploadDate: '1000' } }
    })
    const repo = makeRepo({ getSourceMangaChapterSummary })
    const uc = new SearchSourceWithChaptersUseCase(repo)
    const res = await uc.execute({ sourceId: 's1', query: '', page: 1, type: 'search' })

    expect(res.items[0]).toMatchObject({ id: '1', chapterCount: null, lastChapter: null })
    expect(res.items[1]).toMatchObject({ id: '2', chapterCount: 20 })
  })

  it('passes the browse type through to searchSource', async () => {
    const repo = makeRepo()
    const uc = new SearchSourceWithChaptersUseCase(repo)
    await uc.execute({ sourceId: 's1', query: '', page: 1, type: 'popular' })

    expect(repo.searchSource).toHaveBeenCalledWith({ sourceId: 's1', query: '', page: 1, type: 'popular' })
  })
})
