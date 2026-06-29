// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it, vi } from 'vitest'
import { FETCH_CHAPTERS, MANGA_EXISTS } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue.operations'
import { GraphqlSuwayomiCatalogueRepository } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'

interface Responses {
  exists: { mangas: { totalCount: number } }
  chapters?: unknown
}

function makeClient(responses: Responses): { client: never, execute: ReturnType<typeof vi.fn> } {
  const execute = vi.fn((doc: unknown) => {
    if (doc === MANGA_EXISTS) {
      return Promise.resolve(responses.exists)
    }

    if (doc === FETCH_CHAPTERS) {
      return Promise.resolve(responses.chapters)
    }

    throw new Error('unexpected operation')
  })

  return { client: { execute } as never, execute }
}

describe('graphqlSuwayomiCatalogueRepository.getMangaChapterSummary', () => {
  it('returns undefined and skips the scrape when the manga does not exist', async () => {
    const { client, execute } = makeClient({ exists: { mangas: { totalCount: 0 } } })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    expect(await repo.getMangaChapterSummary('999')).toBeUndefined()
    // The existence probe ran; the expensive fetchChapters scrape never did.
    expect(execute).toHaveBeenCalledTimes(1)
    expect(execute).toHaveBeenCalledWith(MANGA_EXISTS, { id: 999 })
  })

  it('scrapes and summarises chapters when the manga exists', async () => {
    const { client } = makeClient({
      exists: { mangas: { totalCount: 1 } },
      chapters: { fetchChapters: { chapters: [
        { name: 'Ch.1', uploadDate: '1000' },
        { name: 'Ch.2', uploadDate: '2000' },
      ] } },
    })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    expect(await repo.getMangaChapterSummary('44')).toEqual({
      chapterCount: 2,
      lastChapter: { name: 'Ch.2', uploadedAt: '2000' },
    })
  })

  it('degrades an existing manga to 0 chapters when the source is unavailable', async () => {
    const { client } = makeClient({ exists: { mangas: { totalCount: 1 } }, chapters: { fetchChapters: null } })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    expect(await repo.getMangaChapterSummary('44')).toEqual({ chapterCount: 0, lastChapter: null })
  })
})
