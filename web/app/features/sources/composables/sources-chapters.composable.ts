// SPDX-License-Identifier: AGPL-3.0-or-later

import type { SourceComicChapter } from '../types'
import type { ComicSource } from '~/features/comics/types'
import { useSourceChaptersStore } from '../stores/sources-chapters.store'
import { createSourceApi } from './sources.api'

export interface SourceChaptersComposable {
  chapters: Ref<SourceComicChapter[]>
  loading: Ref<boolean>
  sort: Ref<'asc' | 'desc'>
  fetchChapters: (slug: string) => Promise<void>
}

export function useSourceChapters(sourceId: ComicSource): SourceChaptersComposable {
  const api = createSourceApi(sourceId)
  const store = useSourceChaptersStore(sourceId)
  const { t } = useI18n()
  const toast = useToast()

  const loading = ref(false)
  const sort = ref<'asc' | 'desc'>('desc')

  async function fetchChapters(slug: string): Promise<void> {
    loading.value = true

    const res = await api.getSeriesChapters(slug)
    if (res.success) {
      store.setChapters(res.data)
    } else {
      console.error('api.getSeriesChapters', res.error)
      toast.error(t('sources.comic.chapters.error.fetch'))
    }

    loading.value = false
  }

  const chapters = computed(() => {
    return store.chapters.toSorted((a, b) => {
      return sort.value === 'asc' ? a.number - b.number : b.number - a.number
    })
  })

  return {
    chapters,
    loading,
    sort,
    fetchChapters,
  }
}
