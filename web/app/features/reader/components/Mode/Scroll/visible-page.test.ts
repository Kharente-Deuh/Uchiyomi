// SPDX-License-Identifier: AGPL-3.0-or-later

import { describe, expect, it } from 'vitest'
import { pageFromVisibleIndices, recordIntersection } from './visible-page'

describe('recordIntersection', () => {
  it('adds and removes indices', () => {
    const indices = new Set<number>()

    recordIntersection(indices, 2, true)
    recordIntersection(indices, 3, true)
    expect(indices).toEqual(new Set([2, 3]))

    recordIntersection(indices, 2, false)
    expect(indices).toEqual(new Set([3]))
  })
})

describe('pageFromVisibleIndices', () => {
  it('returns the topmost visible page when following the scroll', () => {
    expect(pageFromVisibleIndices(new Set([2, 3, 4]), { movingToPage: null })).toBe(2)
  })

  it('returns undefined when nothing is visible', () => {
    expect(pageFromVisibleIndices(new Set(), { movingToPage: null })).toBeUndefined()
  })

  it('ignores other visible pages until the page we are moving to is on screen', () => {
    expect(pageFromVisibleIndices(new Set([0, 1, 2]), { movingToPage: 10 })).toBeUndefined()
  })

  it('returns the target page once it is visible, even if leftover pages are still listed', () => {
    expect(pageFromVisibleIndices(new Set([0, 1, 10, 11]), { movingToPage: 10 })).toBe(10)
  })

  it('returns the target page when jumping backward past leftover later pages', () => {
    expect(pageFromVisibleIndices(new Set([5, 20, 40]), { movingToPage: 5 })).toBe(5)
  })
})
