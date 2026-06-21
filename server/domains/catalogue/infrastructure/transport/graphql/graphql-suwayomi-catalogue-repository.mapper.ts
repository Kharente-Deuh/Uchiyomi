// SPDX-License-Identifier: AGPL-3.0-or-later
import * as Chapter from '../../../chapter.domain'
import * as Manga from '../../../manga.domain'
import * as Source from '../../../source.domain'

// SourceNode mirrors the GraphQL SourceType selection from LIST_SOURCES.
export interface SourceNode {
  id: string
  name: string
  lang: string
  iconUrl: string
  isNsfw: boolean
}

export function sourceToDomain(node: SourceNode): Source.Model {
  return new Source.Model({
    id: node.id,
    name: node.name,
    lang: node.lang,
    iconUrl: node.iconUrl,
    isNsfw: node.isNsfw,
  })
}

// NOTE: MangaType.id is Int! in the SDL — codegen maps it to number.
// Domain Manga.Summary.id is string; we convert with String().
export interface MangaSummaryNode {
  id: number
  title: string
  thumbnailUrl: string | null
  inLibrary: boolean
}

// NOTE: ChapterType.id is Int!, chapterNumber is Float! (both map to number).
// uploadDate is LongString! (non-nullable in SDL) — domain type matches: string.
export interface ChapterNode {
  id: number
  name: string
  chapterNumber: number
  uploadDate: string
  isDownloaded: boolean
}

export interface MangaDetailsNode extends MangaSummaryNode {
  author: string | null
  description: string | null
  status: string
  chapters: { nodes: ChapterNode[] }
}

export function mangaSummaryToDomain(node: MangaSummaryNode): Manga.Summary {
  return new Manga.Summary({
    id: String(node.id),
    title: node.title,
    thumbnailUrl: node.thumbnailUrl ?? undefined,
    inLibrary: node.inLibrary,
  })
}

export function chapterToDomain(node: ChapterNode): Chapter.Model {
  return new Chapter.Model({
    id: String(node.id),
    name: node.name,
    chapterNumber: node.chapterNumber,
    uploadDate: node.uploadDate,
    isDownloaded: node.isDownloaded,
  })
}

export function mangaDetailsToDomain(node: MangaDetailsNode): Manga.Details {
  return new Manga.Details({
    id: String(node.id),
    title: node.title,
    thumbnailUrl: node.thumbnailUrl ?? undefined,
    inLibrary: node.inLibrary,
    author: node.author ?? undefined,
    description: node.description ?? undefined,
    status: node.status,
    chapters: node.chapters.nodes.map(n => chapterToDomain(n)),
  })
}
