// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { filtersRequireSearch, parseSourceFiltersParam, searchQueryMissing } from '../../server/domains/catalogue/infrastructure/transport/http/parse-source-filters'

describe('parseSourceFiltersParam', () => {
  it('returns [] when the param is absent or empty', () => {
    expect(parseSourceFiltersParam()).toEqual([])
    expect(parseSourceFiltersParam('')).toEqual([])
  })

  it('parses a valid JSON array of changes', () => {
    const raw = JSON.stringify([{ position: 0, checkBoxState: true }, { position: 1, groupChange: { position: 2, triState: 'INCLUDE' } }])
    expect(parseSourceFiltersParam(raw)).toEqual([
      { position: 0, checkBoxState: true },
      { position: 1, groupChange: { position: 2, triState: 'INCLUDE' } },
    ])
  })

  it('returns null for malformed JSON', () => {
    expect(parseSourceFiltersParam('{not json')).toBeNull()
  })

  it('returns null for a valid-JSON but wrong-shape payload', () => {
    expect(parseSourceFiltersParam(JSON.stringify([{ position: 'x' }]))).toBeNull()
    expect(parseSourceFiltersParam(JSON.stringify([{ position: 0, triState: 'MAYBE' }]))).toBeNull()
  })
})

describe('searchQueryMissing', () => {
  it('is true only for a search with neither a query nor filters', () => {
    expect(searchQueryMissing('search', '', [])).toBe(true)
    expect(searchQueryMissing('search', 'naruto', [])).toBe(false)
    expect(searchQueryMissing('search', '', [{ position: 0, checkBoxState: true }])).toBe(false)
    expect(searchQueryMissing('popular', '', [])).toBe(false)
    expect(searchQueryMissing('latest', '', [])).toBe(false)
  })
})

describe('filtersRequireSearch', () => {
  const oneChange = [{ position: 0, checkBoxState: true }]

  it('is true for popular with filters', () => {
    expect(filtersRequireSearch('popular', oneChange)).toBe(true)
  })

  it('is true for latest with filters', () => {
    expect(filtersRequireSearch('latest', oneChange)).toBe(true)
  })

  it('is false for search with filters', () => {
    expect(filtersRequireSearch('search', oneChange)).toBe(false)
  })

  it('is false for popular with no filters', () => {
    expect(filtersRequireSearch('popular', [])).toBe(false)
  })

  it('is false for latest with no filters', () => {
    expect(filtersRequireSearch('latest', [])).toBe(false)
  })

  it('is false for search with no filters', () => {
    expect(filtersRequireSearch('search', [])).toBe(false)
  })
})
