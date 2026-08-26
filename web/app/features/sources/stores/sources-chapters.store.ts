// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceComicChapter } from '../types'
import type { ComicSource } from '~/features/comics/types'

export interface SourceChaptersStore {
  chapters: Ref<SourceComicChapter[]>
  setChapters: (value: SourceComicChapter[]) => void
  invalidate: () => void
}

type ChaptersStoreDefinition = ReturnType<typeof defineStore<string, SourceChaptersStore>>

const chaptersStoresMap = new Map<ComicSource, ChaptersStoreDefinition>()

function createChaptersStoreDefinition(sourceId: ComicSource): ChaptersStoreDefinition {
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

export function useSourceChaptersStore(sourceId: ComicSource): ReturnType<ChaptersStoreDefinition> {
  if (!chaptersStoresMap.has(sourceId)) {
    chaptersStoresMap.set(sourceId, createChaptersStoreDefinition(sourceId))
  }

  return chaptersStoresMap.get(sourceId)!()
}
