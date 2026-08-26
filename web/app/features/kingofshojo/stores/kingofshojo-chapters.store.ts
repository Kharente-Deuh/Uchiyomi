// SPDX-License-Identifier: AGPL-3.0-or-later

import type { KingOfShojoComicChapter } from '../types'

export interface KingOfShojoChaptersStore {
  chapters: KingOfShojoComicChapter[]
  setChapters: (value: KingOfShojoComicChapter[]) => void
  invalidate: () => void
}

export const useKingOfShojoChaptersStore = defineStore('kingofshojo-chapters', () => {
  const chapters = ref<KingOfShojoComicChapter[]>([])

  function setChapters(value: KingOfShojoComicChapter[]): void {
    chapters.value = value
  }

  function invalidate(): void {
    chapters.value = []
  }

  return {
    chapters,

    setChapters,
    invalidate,
  }
})
