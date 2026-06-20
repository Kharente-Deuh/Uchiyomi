// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment node
import { describe, expect, it } from 'vitest'
import { mapChapter, mapMangaDetails, mapMangaSummary, mapSource } from '../../server/domains/catalogue/infrastructure/mappers'

describe('mapSource', () => {
  it('maps a Suwayomi source node to the domain Source', () => {
    const node = { id: '42', name: 'MangaDex', lang: 'en', iconUrl: '/icon.png', isNsfw: false }
    expect(mapSource(node)).toEqual({
      id: '42',
      name: 'MangaDex',
      lang: 'en',
      iconUrl: '/icon.png',
      isNsfw: false,
    })
  })
})

describe('mapMangaSummary', () => {
  it('maps a summary node and normalises a missing thumbnail to null', () => {
    // NOTE: SDL MangaType.id is Int! — infrastructure node uses number; domain id is string.
    expect(mapMangaSummary({ id: 1, title: 'T', thumbnailUrl: null, inLibrary: false })).toEqual({
      id: '1',
      title: 'T',
      thumbnailUrl: null,
      inLibrary: false,
    })
  })
})

describe('mapChapter', () => {
  it('maps a chapter node', () => {
    // NOTE: SDL ChapterType.id is Int!, chapterNumber is Float!, uploadDate is LongString! (non-nullable).
    expect(mapChapter({
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

describe('mapMangaDetails', () => {
  it('maps details with its chapters and normalises nullables', () => {
    const result = mapMangaDetails({
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
    expect(result.author).toBeNull()
    expect(result.chapters).toHaveLength(1)
    expect(result.chapters[0].id).toBe('9')
  })
})
