// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import process from 'node:process'
import { describe, expect, it } from 'vitest'
import * as GetMangaDetails from '../../server/domains/catalogue/application/usecases/get-manga-details.use-case'
import * as ListSources from '../../server/domains/catalogue/application/usecases/list-sources.use-case'
import * as SearchSource from '../../server/domains/catalogue/application/usecases/search-source.use-case'
import { GraphqlSuwayomiCatalogueRepository } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'
import { createSuwayomiClient } from '../../server/utils/suwayomi/client'

const base = process.env.SUWAYOMI_URL
const testSourceId = process.env.SUWAYOMI_TEST_SOURCE_ID
const testMangaId = process.env.SUWAYOMI_TEST_MANGA_ID
const describeIf = base ? describe : describe.skip

describeIf('catalogue integration (live Suwayomi)', () => {
  const repo = new GraphqlSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )

  it('lists sources from the live server', async () => {
    const sources = await new ListSources.ListSourcesUseCase(repo).execute()
    expect(Array.isArray(sources)).toBe(true)
    // A fresh server may have zero sources; assert the shape is well-formed when present.
    for (const s of sources) {
      expect(typeof s.id).toBe('string')
    }
  })
})

// These need an installed extension + a live external site, so they run only when
// a test source/manga is configured locally; otherwise they skip.
const describeSearch = base && testSourceId ? describe : describe.skip
describeSearch('catalogue search (live source)', () => {
  const repo = new GraphqlSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )
  it('returns a well-formed search page', async () => {
    const result = await new SearchSource.SearchSourceUseCase(repo).execute({ sourceId: testSourceId!, query: 'a', page: 1, type: 'search' })
    expect(Array.isArray(result.mangas)).toBe(true)
    expect(typeof result.hasNextPage).toBe('boolean')
  })
})

const describeDetails = base && testMangaId ? describe : describe.skip
describeDetails('catalogue manga details (live)', () => {
  const repo = new GraphqlSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )
  it('returns details with chapters', async () => {
    const details = await new GetMangaDetails.GetMangaDetailsUseCase(repo).execute({ mangaId: testMangaId! })
    expect(typeof details.title).toBe('string')
    expect(Array.isArray(details.chapters)).toBe(true)
  })
})
