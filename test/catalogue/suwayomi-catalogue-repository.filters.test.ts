// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import type { SuwayomiClient } from '../../server/utils/suwayomi/client'
import { describe, expect, it, vi } from 'vitest'
import { GraphqlSuwayomiCatalogueRepository } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue.repository'

function fakeClient(result: unknown): { client: SuwayomiClient, execute: ReturnType<typeof vi.fn> } {
  const execute = vi.fn().mockResolvedValue(result)

  return { client: { execute } as unknown as SuwayomiClient, execute }
}

describe('graphqlSuwayomiCatalogueRepository.getSourceFilters', () => {
  it('maps the source filter union to domain filters', async () => {
    const { client } = fakeClient({
      source: { filters: [{ __typename: 'CheckBoxFilter', name: 'Completed', checkBoxDefault: true }] },
    })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    expect(await repo.getSourceFilters('s1')).toEqual([
      { type: 'checkbox', position: 0, name: 'Completed', default: true },
    ])
  })
})

describe('graphqlSuwayomiCatalogueRepository.searchSource with filters', () => {
  it('passes mapped filter changes only for the search type', async () => {
    const { client, execute } = fakeClient({ fetchSourceManga: { mangas: [], hasNextPage: false } })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    await repo.searchSource({ sourceId: 's1', query: 'q', page: 1, type: 'search', filters: [{ position: 0, checkBoxState: true }] })

    expect(execute).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({
      filters: [{ position: 0, checkBoxState: true }],
    }))
  })

  it('omits filters for popular browse', async () => {
    const { client, execute } = fakeClient({ fetchSourceManga: { mangas: [], hasNextPage: false } })
    const repo = new GraphqlSuwayomiCatalogueRepository(client)

    await repo.searchSource({ sourceId: 's1', query: '', page: 1, type: 'popular', filters: [{ position: 0, checkBoxState: true }] })

    expect(execute).toHaveBeenCalledWith(expect.anything(), expect.objectContaining({ filters: undefined }))
  })
})
