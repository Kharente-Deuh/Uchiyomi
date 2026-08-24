// SPDX-License-Identifier: AGPL-3.0-or-later

import type { AsuraScansComicChapter } from '../types'

export interface AsuraScansChaptersStore {
  chapters: AsuraScansComicChapter[]
  setChapters: (value: AsuraScansComicChapter[]) => void
  invalidate: () => void
}

export const useAsuraScansChaptersStore = defineStore('asurascansChapters', () => {
  const chapters = ref<AsuraScansComicChapter[]>([])

  function setChapters(value: AsuraScansComicChapter[]): void {
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
