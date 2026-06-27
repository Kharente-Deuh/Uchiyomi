// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { chapterSummaryFromFetched } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue-repository.mapper'

describe('chapterSummaryFromFetched', () => {
  it('counts chapters and picks the most recently uploaded as lastChapter', () => {
    const summary = chapterSummaryFromFetched([
      { name: 'Chapter 1', uploadDate: '1000' },
      { name: 'Chapter 3', uploadDate: '3000' },
      { name: 'Chapter 2', uploadDate: '2000' },
    ])

    expect(summary.chapterCount).toBe(3)
    expect(summary.lastChapter).toEqual({ name: 'Chapter 3', uploadDate: '3000' })
  })

  it('returns count 0 and null lastChapter when there are no chapters', () => {
    expect(chapterSummaryFromFetched([])).toEqual({ chapterCount: 0, lastChapter: null })
  })
})
