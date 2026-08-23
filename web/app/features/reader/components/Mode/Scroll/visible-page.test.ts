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
  it('returns the topmost visible page', () => {
    expect(pageFromVisibleIndices(new Set([2, 3, 4]), { restoredPage: 0, restoring: false })).toBe(2)
  })

  it('returns undefined when nothing is visible', () => {
    expect(pageFromVisibleIndices(new Set(), { restoredPage: 0, restoring: false })).toBeUndefined()
  })

  it('ignores pages above the restored position while restoring progress', () => {
    expect(pageFromVisibleIndices(new Set([0, 1, 2]), { restoredPage: 10, restoring: true })).toBeUndefined()
  })

  it('uses the restored window once that page is on screen, ignoring leftover top pages', () => {
    expect(pageFromVisibleIndices(new Set([0, 1, 10, 11]), { restoredPage: 10, restoring: true })).toBe(10)
  })
})
