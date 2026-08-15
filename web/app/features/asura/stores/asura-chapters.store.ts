import type { AsuraComicChapter } from '../types'

export interface AsuraChaptersStore {
  chapters: AsuraComicChapter[]
  setChapters: (value: AsuraComicChapter[]) => void
  invalidate: () => void
}

export const useAsuraChaptersStore = defineStore('asuraChapters', () => {
  const chapters = ref<AsuraComicChapter[]>([])

  function setChapters(value: AsuraComicChapter[]): void {
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
