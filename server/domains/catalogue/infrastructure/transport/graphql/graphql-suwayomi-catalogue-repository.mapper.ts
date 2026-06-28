// SPDX-License-Identifier: AGPL-3.0-or-later

import type { FilterChangeInput } from '../../../../../utils/suwayomi/generated/graphql'
import type { SourceFilter, SourceFilterChange } from '../../../catalogue.domain'
import { ChapterModel } from '../../../chapter.domain'
import { MangaChapterSummaryModel, MangaDetailsModel, MangaSummaryModel } from '../../../manga.domain'
import { SourceModel } from '../../../source.domain'

// SourceNode mirrors the GraphQL SourceType selection from LIST_SOURCES.
export interface SourceNode {
  id: string
  name: string
  lang: string
  iconUrl: string
  isNsfw: boolean
}

export function sourceToDomain(node: SourceNode): SourceModel {
  return new SourceModel({
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
  realUrl: string | null
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

// GET_MANGA_DETAILS does not select realUrl, so it is omitted here; mangaDetailsToDomain
// leaves the domain model's realUrl undefined.
export interface MangaDetailsNode extends Omit<MangaSummaryNode, 'realUrl'> {
  author: string | null
  description: string | null
  status: string
  chapters: { nodes: ChapterNode[] }
}

export function mangaSummaryToDomain(node: MangaSummaryNode): MangaSummaryModel {
  return new MangaSummaryModel({
    id: String(node.id),
    title: node.title,
    thumbnailUrl: node.thumbnailUrl ?? undefined,
    inLibrary: node.inLibrary,
    realUrl: node.realUrl ?? undefined,
  })
}

export function chapterToDomain(node: ChapterNode): ChapterModel {
  return new ChapterModel({
    id: String(node.id),
    name: node.name,
    chapterNumber: node.chapterNumber,
    uploadDate: node.uploadDate,
    isDownloaded: node.isDownloaded,
  })
}

export function mangaDetailsToDomain(node: MangaDetailsNode): MangaDetailsModel {
  return new MangaDetailsModel({
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

// FetchChaptersPayload.chapters selection — name + uploadDate only.
export interface FetchedChapterNode {
  name: string
  uploadDate: string
}

// Derive a chapter summary from a source's freshly-fetched chapter list.
// uploadDate is LongString! epoch ms; "last" is the max by upload date.
export function chapterSummaryFromFetched(chapters: FetchedChapterNode[]): MangaChapterSummaryModel {
  if (chapters.length === 0) {
    return { chapterCount: 0, lastChapter: null }
  }

  let latest = chapters[0]!
  for (const c of chapters) {
    if (Number(c.uploadDate) > Number(latest.uploadDate)) {
      latest = c
    }
  }

  return new MangaChapterSummaryModel({
    chapterCount: chapters.length,
    lastChapter: latest ? { name: latest.name, uploadedAt: latest.uploadDate } : null,
  })
}

// Structural mirror of the GET_SOURCE_FILTERS selection (decoupled from codegen
// names, matching the existing *Node convention in this file). __typename
// discriminates the Suwayomi `Filter` union; `default` is aliased per member.
export type FilterLeafNode
  = | { __typename: 'CheckBoxFilter', name: string, checkBoxDefault: boolean }
    | { __typename: 'TriStateFilter', name: string, triDefault: 'IGNORE' | 'INCLUDE' | 'EXCLUDE' }
    | { __typename: 'SelectFilter', name: string, selectDefault: number, values: string[] }
    | { __typename: 'TextFilter', name: string, textDefault: string }
    | { __typename: 'SortFilter', name: string, sortDefault: { ascending: boolean, index: number } | null, values: string[] }
    | { __typename: 'HeaderFilter', name: string }
    | { __typename: 'SeparatorFilter', name: string }

export type FilterNode
  = | FilterLeafNode
    | { __typename: 'GroupFilter', name: string, filters: FilterLeafNode[] }

function leafFilterToDomain(node: FilterLeafNode, position: number): SourceFilter {
  switch (node.__typename) {
    case 'CheckBoxFilter': return { type: 'checkbox', position, name: node.name, default: node.checkBoxDefault }
    case 'TriStateFilter': return { type: 'tristate', position, name: node.name, default: node.triDefault }
    case 'SelectFilter': return { type: 'select', position, name: node.name, default: node.selectDefault, values: node.values }
    case 'TextFilter': return { type: 'text', position, name: node.name, default: node.textDefault }
    case 'SortFilter': return { type: 'sort', position, name: node.name, default: node.sortDefault, values: node.values }
    case 'HeaderFilter': return { type: 'header', position, name: node.name }
    case 'SeparatorFilter': return { type: 'separator', position, name: node.name }
  }
}

export function sourceFiltersToDomain(nodes: FilterNode[]): SourceFilter[] {
  return nodes.map((node, position) => {
    if (node.__typename === 'GroupFilter') {
      return {
        type: 'group',
        position,
        name: node.name,
        filters: node.filters.map((child, childIndex) => leafFilterToDomain(child, childIndex)),
      }
    }

    return leafFilterToDomain(node, position)
  })
}

export function filterChangeToInput(change: SourceFilterChange): FilterChangeInput {
  return {
    position: change.position,
    ...(change.checkBoxState !== undefined && { checkBoxState: change.checkBoxState }),
    ...(change.triState !== undefined && { triState: change.triState }),
    ...(change.selectState !== undefined && { selectState: change.selectState }),
    ...(change.textState !== undefined && { textState: change.textState }),
    ...(change.sortState !== undefined && { sortState: change.sortState }),
    ...(change.groupChange !== undefined && { groupChange: filterChangeToInput(change.groupChange) }),
  }
}
