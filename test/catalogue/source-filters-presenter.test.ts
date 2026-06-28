// SPDX-License-Identifier: AGPL-3.0-or-later

// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { toSourceFiltersDto } from '../../server/domains/catalogue/infrastructure/transport/http/catalogue-http.presenter'

describe('toSourceFiltersDto', () => {
  it('passes leaf filters through unchanged', () => {
    expect(toSourceFiltersDto([
      { type: 'checkbox', position: 0, name: 'Completed', default: false },
      { type: 'sort', position: 1, name: 'Sort', default: { ascending: true, index: 0 }, values: ['Title'] },
    ])).toEqual([
      { type: 'checkbox', position: 0, name: 'Completed', default: false },
      { type: 'sort', position: 1, name: 'Sort', default: { ascending: true, index: 0 }, values: ['Title'] },
    ])
  })

  it('recurses into groups', () => {
    expect(toSourceFiltersDto([
      { type: 'group', position: 0, name: 'Genres', filters: [{ type: 'tristate', position: 0, name: 'Comedy', default: 'IGNORE' }] },
    ])).toEqual([
      { type: 'group', position: 0, name: 'Genres', filters: [{ type: 'tristate', position: 0, name: 'Comedy', default: 'IGNORE' }] },
    ])
  })
})
