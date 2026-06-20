// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import process from 'node:process'
import { describe, expect, it } from 'vitest'
import { getMangaDetails } from '../../server/domains/catalogue/application/getMangaDetails'
import { listSources } from '../../server/domains/catalogue/application/listSources'
import { searchSource } from '../../server/domains/catalogue/application/searchSource'
import { createSuwayomiCatalogueRepository } from '../../server/domains/catalogue/infrastructure/suwayomiCatalogueRepository'
import { createSuwayomiClient } from '../../server/utils/suwayomi/client'

const base = process.env.SUWAYOMI_URL
const testSourceId = process.env.SUWAYOMI_TEST_SOURCE_ID
const testMangaId = process.env.SUWAYOMI_TEST_MANGA_ID
const describeIf = base ? describe : describe.skip

describeIf('catalogue integration (live Suwayomi)', () => {
  const repo = createSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )

  it('lists sources from the live server', async () => {
    const sources = await listSources(repo)
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
  const repo = createSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )
  it('returns a well-formed search page', async () => {
    const result = await searchSource(repo, testSourceId!, 'a', 1)
    expect(Array.isArray(result.mangas)).toBe(true)
    expect(typeof result.hasNextPage).toBe('boolean')
  })
})

const describeDetails = base && testMangaId ? describe : describe.skip
describeDetails('catalogue manga details (live)', () => {
  const repo = createSuwayomiCatalogueRepository(
    createSuwayomiClient({ endpoint: `${base}/api/graphql` }),
  )
  it('returns details with chapters', async () => {
    const details = await getMangaDetails(repo, testMangaId!)
    expect(typeof details.title).toBe('string')
    expect(Array.isArray(details.chapters)).toBe(true)
  })
})
