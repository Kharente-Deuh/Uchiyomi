// SPDX-License-Identifier: AGPL-3.0-or-later

export type ComicStatus = 'ongoing' | 'completed' | 'hiatus' | 'cancelled'
export type ComicType = 'manga' | 'manhua' | 'manhwa' | 'mangatoon'

export interface LightComic {
  id: string
  slug: string
  source: string
  status: ComicStatus
  chapterCount: number
}
