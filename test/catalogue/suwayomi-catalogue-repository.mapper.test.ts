// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { chapterToDomain, mangaDetailsToDomain, mangaSummaryToDomain, sourceToDomain } from '../../server/domains/catalogue/infrastructure/transport/graphql/graphql-suwayomi-catalogue-repository.mapper'

describe('sourceToDomain', () => {
  it('maps a Suwayomi source node to the domain Source', () => {
    const node = { id: '42', name: 'MangaDex', lang: 'en', iconUrl: '/icon.png', isNsfw: false }
    expect(sourceToDomain(node)).toEqual({
      id: '42',
      name: 'MangaDex',
      lang: 'en',
      iconUrl: '/icon.png',
      isNsfw: false,
    })
  })
})

describe('mangaSummaryToDomain', () => {
  it('maps a summary node and normalises a missing thumbnail to undefined', () => {
    // NOTE: SDL MangaType.id is Int! — infrastructure node uses number; domain id is string.
    const result = mangaSummaryToDomain({ id: 1, title: 'T', thumbnailUrl: null, inLibrary: false })
    expect(result.id).toBe('1')
    expect(result.title).toBe('T')
    expect(result.thumbnailUrl).toBeUndefined()
    expect(result.inLibrary).toBe(false)
  })
})

describe('chapterToDomain', () => {
  it('maps a chapter node', () => {
    // NOTE: SDL ChapterType.id is Int!, chapterNumber is Float!, uploadDate is LongString! (non-nullable).
    expect(chapterToDomain({
      id: 9,
      name: 'Ch 9',
      chapterNumber: 9,
      uploadDate: '1700000000000',
      isDownloaded: true,
    })).toEqual({
      id: '9',
      name: 'Ch 9',
      chapterNumber: 9,
      uploadDate: '1700000000000',
      isDownloaded: true,
    })
  })
})

describe('mangaDetailsToDomain', () => {
  it('maps details with its chapters and normalises nullables', () => {
    const result = mangaDetailsToDomain({
      id: 1,
      title: 'T',
      thumbnailUrl: null,
      inLibrary: true,
      author: null,
      description: null,
      status: 'ONGOING',
      chapters: { nodes: [{ id: 9, name: 'Ch 9', chapterNumber: 9, uploadDate: '1700000000000', isDownloaded: false }] },
    })
    expect(result.id).toBe('1')
    expect(result.author).toBeUndefined()
    expect(result.chapters).toHaveLength(1)
    expect(result.chapters[0].id).toBe('9')
  })
})
