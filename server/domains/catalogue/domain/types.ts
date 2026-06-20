// SPDX-License-Identifier: AGPL-3.0-or-later
export interface Source {
  id: string
  name: string
  lang: string
  iconUrl: string
  isNsfw: boolean
}

export interface MangaSummary {
  id: string
  title: string
  thumbnailUrl: string | null
  inLibrary: boolean
}

export interface Chapter {
  id: string
  name: string
  chapterNumber: number
  uploadDate: string
  isDownloaded: boolean
}

export interface MangaDetails extends MangaSummary {
  author: string | null
  description: string | null
  status: string
  chapters: Chapter[]
}

export interface SearchResult {
  mangas: MangaSummary[]
  hasNextPage: boolean
}
