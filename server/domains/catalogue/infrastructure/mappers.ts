// SPDX-License-Identifier: AGPL-3.0-or-later
import type { Chapter, MangaDetails, MangaSummary, Source } from '../domain/types'

// SourceNode mirrors the GraphQL SourceType selection from LIST_SOURCES.
export interface SourceNode {
  id: string
  name: string
  lang: string
  iconUrl: string
  isNsfw: boolean
}

export function mapSource(node: SourceNode): Source {
  return {
    id: node.id,
    name: node.name,
    lang: node.lang,
    iconUrl: node.iconUrl,
    isNsfw: node.isNsfw,
  }
}

// NOTE: MangaType.id is Int! in the SDL — codegen maps it to number.
// Domain MangaSummary.id is string; we convert with String().
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

export function mapMangaSummary(node: MangaSummaryNode): MangaSummary {
  return {
    id: String(node.id),
    title: node.title,
    thumbnailUrl: node.thumbnailUrl ?? null,
    inLibrary: node.inLibrary,
  }
}

export function mapChapter(node: ChapterNode): Chapter {
  return {
    id: String(node.id),
    name: node.name,
    chapterNumber: node.chapterNumber,
    uploadDate: node.uploadDate,
    isDownloaded: node.isDownloaded,
  }
}

export function mapMangaDetails(node: MangaDetailsNode): MangaDetails {
  return {
    ...mapMangaSummary(node),
    author: node.author ?? null,
    description: node.description ?? null,
    status: node.status,
    chapters: node.chapters.nodes.map(n => mapChapter(n)),
  }
}
