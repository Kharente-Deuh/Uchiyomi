<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicProgressContinue } from '../../types'
import type { Chapter } from '~/features/chapters/types'

const props = defineProps<{
  id: string
  continue?: ComicProgressContinue
}>()

const POLL_INTERVAL_MS = 2000

const { t } = useI18n()
const toast = useToast()
const { smAndDown } = useDisplay()
const api = createChaptersApi()

const chapters = ref<Chapter[]>([])
const loading = ref(false)
const sort = ref<'asc' | 'desc'>('desc')

const sortedChapters = computed(() => chapters.value.toSorted((a, b) => sort.value === 'asc'
  ? a.number - b.number
  : b.number - a.number))

async function fetchChapters(): Promise<void> {
  loading.value = true
  const res = await api.getByComicId(props.id)
  if (!res.success) {
    console.error('api.getChapters failed', res.error)
    toast.error(res.error.status === 404
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    loading.value = false

    return
  }

  chapters.value = res.data
  loading.value = false

  syncPolling()
}

const { pause, resume } = useIntervalFn(() => {
  void pollChapters()
}, POLL_INTERVAL_MS, { immediate: false })

function syncPolling(): void {
  if (chapters.value.some(chapter => chapter.download >= 0 && chapter.download < 100)) {
    resume()
  } else {
    pause()
  }
}

const isInFlight = ref(false)
async function pollChapters(): Promise<void> {
  if (isInFlight.value || loading.value) {
    return
  }

  const chaptersIds: string[] = []
  for (const chapter of chapters.value) {
    if (chapter.download >= 0 && chapter.download < 100) {
      chaptersIds.push(chapter.id)
    }
  }

  isInFlight.value = true

  const res = await api.getByIds(chaptersIds)
  if (res.success) {
    for (const chapter of res.data) {
      const i = chapters.value.findIndex(c => c.id === chapter.id)
      if (i === -1) {
        continue
      }

      if (chapter.download === chapters.value[i]!.download) {
        continue
      }

      chapters.value[i] = chapter
    }
  }

  isInFlight.value = false
  syncPolling()
}

watch(() => props.id, () => {
  fetchChapters()
}, { immediate: true })

const retryChaptersLoading = ref<Record<string, boolean>>({})

async function retryChapter(chapterId: string): Promise<void> {
  if (retryChaptersLoading.value[chapterId]) {
    return
  }

  retryChaptersLoading.value[chapterId] = true

  const res = await api.retryDownload(chapterId)
  if (res.success) {
    await fetchChapters()

    delete retryChaptersLoading.value[chapterId]

    return
  }

  console.error('chaptersApi.retryDownload', res.error)

  switch (res.error.status) {
    case 404:
      toast.error(t('sources.asura.comic.chapters.error.retry.notFound'))
      break
    case 403:
      toast.error(t('sources.asura.comic.chapters.error.retry.forbidden'))
      break
    case 409:
      toast.error(t('sources.asura.comic.chapters.error.retry.conflict'))
      break
    default:
      toast.error(t('error.unknown'))
      break
  }

  delete retryChaptersLoading.value[chapterId]
}

const nextChapter = computed(() => {
  const chaptersCpy = sort.value === 'asc' ? chapters.value : chapters.value.toSorted((a, b) => a.number - b.number)
  const continueCpy = props.continue

  if (!continueCpy) {
    return chaptersCpy[0] as Chapter
  }

  const i = chaptersCpy.findIndex(chapter => chapter.id === continueCpy.chapterId)
  if (i === -1) {
    return
  }

  if (continueCpy.page === chaptersCpy[i]!.pagesNb) {
    if (i === chaptersCpy.length - 1) {
      return
    }

    return chaptersCpy[i + 1] as Chapter
  }

  return chaptersCpy[i] as Chapter
})
const nextChapterText = computed(() => {
  if (!nextChapter.value || smAndDown.value) {
    return
  }

  return nextChapter.value.number === 1 ? $t('common.start') : $t('common.continue')
})

const sortIcon = computed(() => sort.value === 'asc' ? 'fa6-solid:arrow-down-short-wide' : 'fa6-solid:arrow-up-short-wide')
const sortText = computed(() => {
  if (smAndDown.value) {
    return
  }

  return sort.value === 'asc' ? $t('common.sort.oldest') : $t('common.sort.latest')
})
</script>

<template>
  <div class="d-flex flex-column w-100 position-relative bg-surface" style="border-radius: 12px; max-height: 40rem;">
    <div class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center" style="z-index: 1; border-top-left-radius: 12px; border-top-right-radius: 12px;">
      <span class="text-title-large font-weight-bold">{{ $t('sources.asurascans.comic.chaptersCount', { count: chapters.length }) }}</span>
      <AtomLink
        v-if="nextChapter"
        :to="nextChapter && nextChapter.download === 100 ? `/comic/${props.id}/${nextChapter.id}` : undefined"
      >
        <VBtn
          variant="tonal"
          class="border-thin-primary"
          :icon="smAndDown ? 'fa6-solid:play' : undefined"
          :prepend-icon="smAndDown ? undefined : 'fa6-solid:play'"
          :text="nextChapterText"
          :size="smAndDown ? 'small' : undefined"
        />
      </AtomLink>
      <VBtn
        variant="tonal"
        class="text-body-medium"
        color="surfaceVariant"
        :icon="smAndDown ? sortIcon : undefined"
        :text="sortText"
        :size="smAndDown ? 'small' : undefined"
        :prepend-icon="smAndDown ? undefined : sortIcon"
        @click="sort = sort === 'asc' ? 'desc' : 'asc'"
      />
    </div>

    <div v-if="loading || chapters.length === 0" class="d-flex justify-center align-center w-100 pa-4">
      <VProgressCircular
        v-if="loading"
        class="m-auto"
        indeterminate
        color="primary"
      />
      <span v-else class="text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: 0 }) }}</span>
    </div>
    <VVirtualScroll :items="sortedChapters" style="border-bottom-left-radius: 12px; border-bottom-right-radius: 12px;">
      <template #default="{ item }">
        <ComicsChaptersItem
          :chapter="item"
          :retry-loading="!!retryChaptersLoading[item.id]"
          @retry="retryChapter(item.id)"
        />
      </template>
    </VVirtualScroll>
  </div>
</template>
