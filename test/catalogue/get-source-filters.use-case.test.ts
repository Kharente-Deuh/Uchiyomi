// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import type { CatalogueRepository, SourceFilter } from '../../server/domains/catalogue/catalogue.domain'
import { describe, expect, it, vi } from 'vitest'
import { GetSourceFiltersUseCase } from '../../server/domains/catalogue/application/usecases/get-source-filters.use-case'

describe('getSourceFiltersUseCase', () => {
  it('delegates to the repository with the source id', async () => {
    const filters: SourceFilter[] = [{ type: 'text', position: 0, name: 'Author', default: '' }]
    const repo = { getSourceFilters: vi.fn().mockResolvedValue(filters) } as unknown as CatalogueRepository

    const result = await new GetSourceFiltersUseCase(repo).execute({ sourceId: 's1' })

    expect(repo.getSourceFilters).toHaveBeenCalledWith('s1')
    expect(result).toBe(filters)
  })
})
