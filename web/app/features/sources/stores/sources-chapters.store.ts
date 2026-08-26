// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ComicSource } from '~/features/comics/types'
import type { SourceComicChapter } from '../types'

export interface SourceChaptersStore {
  chapters: Ref<SourceComicChapter[]>
  setChapters: (value: SourceComicChapter[]) => void
  invalidate: () => void
}

const chaptersStoresMap = new Map<ComicSource, ReturnType<typeof createChaptersStoreDefinition>>()

function createChaptersStoreDefinition(sourceId: ComicSource) {
  return defineStore(`sources-chapters-${sourceId}`, (): SourceChaptersStore => {
    const chapters = ref<SourceComicChapter[]>([])

    function setChapters(value: SourceComicChapter[]): void {
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
}

export function useSourceChaptersStore(sourceId: ComicSource): ReturnType<ReturnType<typeof createChaptersStoreDefinition>> {
  if (!chaptersStoresMap.has(sourceId)) {
    chaptersStoresMap.set(sourceId, createChaptersStoreDefinition(sourceId))
  }
  return chaptersStoresMap.get(sourceId)!()
}
