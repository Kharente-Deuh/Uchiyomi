<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicProgressContinue } from '../../types'
import type { Chapter } from '~/features/chapters/types'

const props = defineProps<{
  id: string
  continue?: ComicProgressContinue
}>()

const emit = defineEmits<{ refetchProgress: [] }>()

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

const sortIcon = computed(() => sort.value === 'asc' ? 'fa6-solid:arrow-down-short-wide' : 'fa6-solid:arrow-up-short-wide')
const sortText = computed(() => {
  if (smAndDown.value) {
    return
  }

  return sort.value === 'asc' ? $t('common.sort.oldest') : $t('common.sort.latest')
})

const updateProgressLoading = ref<Record<string, boolean>>({})
async function updateProgress(chapterId: string, mode: 'read' | 'unread'): Promise<void> {
  if (updateProgressLoading.value[chapterId]) {
    return
  }

  updateProgressLoading.value[chapterId] = true

  let success = false

  switch (mode) {
    case 'read':
      success = await setChapterRead(chapterId)
      break
    case 'unread':
      success = await setChapterUnread(chapterId)
      break
  }

  if (success) {
    emit('refetchProgress')
  }

  delete updateProgressLoading.value[chapterId]
}

async function setChapterRead(chapterId: string): Promise<boolean> {
  const index = chapters.value.findIndex(chapter => chapter.id === chapterId)
  if (index === -1) {
    return false
  }

  const res = await api.saveProgress({
    id: chapterId,
    page: chapters.value[index]!.pagesNb,
  })

  if (!res.success) {
    console.error('chaptersApi.saveProgress', res.error)
    toast.error(t('error.unknown'))

    return false
  }

  chapters.value[index]!.progress = res.data

  return true
}

async function setChapterUnread(chapterId: string): Promise<boolean> {
  const index = chapters.value.findIndex(chapter => chapter.id === chapterId)
  if (index === -1) {
    return false
  }

  const res = await api.deleteProgress(chapterId)
  if (!res.success) {
    console.error('chaptersApi.deleteProgress', res.error)
    toast.error(t('error.unknown'))

    return false
  }

  chapters.value[index]!.progress = undefined

  return true
}

const selectedChapters = ref<Chapter[]>([])
const selectableChapters = computed(() => sortedChapters.value.filter(({ earlyAccessUntil }) => !earlyAccessUntil || earlyAccessUntil < new Date()))
const selectAllChaptersIcon = computed(() => {
  if (!selectedChapters.value.length || selectableChapters.value.length === 0) {
    return 'fa6-regular:square'
  }

  if (selectableChapters.value.length > 0 && selectedChapters.value.length === selectableChapters.value.length) {
    return 'fa6-solid:square-check'
  }

  return 'fa6-solid:square-minus'
})

function toggleSelectAllChapters(): void {
  if (selectedChapters.value.length === selectableChapters.value.length) {
    selectedChapters.value = []

    return
  }

  selectedChapters.value = [...selectableChapters.value]
}

function toggleSelectChapter(chapterId: string): void {
  const chapter = chapters.value.find(chapter => chapter.id === chapterId)
  if (!chapter || (chapter.earlyAccessUntil && chapter.earlyAccessUntil > new Date())) {
    return
  }

  if (selectedChapters.value.some(chapter => chapter.id === chapterId)) {
    selectedChapters.value = selectedChapters.value.filter(chapter => chapter.id !== chapterId)
  } else {
    selectedChapters.value.push(chapters.value.find(chapter => chapter.id === chapterId)!)
  }
}

function onSelectionAction(updateContinue: boolean): void {
  if (updateContinue) {
    emit('refetchProgress')
  }

  fetchChapters()
}
</script>

<template>
  <div class="d-flex flex-column w-100 position-relative bg-surface" style="border-radius: 12px; max-height: 40rem;">
    <div
      v-if="selectedChapters.length === 0"
      class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center"
      style="z-index: 1; border-top-left-radius: 12px; border-top-right-radius: 12px;"
    >
      <span class="text-title-large font-weight-bold">{{ $t('sources.asurascans.comic.chaptersCount', { count: chapters.length }) }}</span>
      <ComicsChaptersContinue
        :comic-id="props.id"
        :continue="props.continue"
        :chapters="sortedChapters"
        :sort
      />

      <div class="d-flex align-center ga-4">
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

        <VIcon
          :icon="selectAllChaptersIcon"
          cursor="pointer"
          size="large"
          @click="toggleSelectAllChapters"
        />
      </div>
    </div>
    <div
      v-else
      class="d-flex ga-6 pa-4 border-b-thin bg-surface align-center"
      style="z-index: 1; border-top-left-radius: 12px; border-top-right-radius: 12px;"
    >
      <div class="d-flex align-center w-100 justify-center">
        <ComicsChaptersActions
          v-model="selectedChapters"
          :comic-id="props.id"
          @refetch-chapters="onSelectionAction"
        />
      </div>

      <VIcon
        :icon="selectAllChaptersIcon"
        cursor="pointer"
        size="large"
        @click="toggleSelectAllChapters"
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
          :disabled="item.earlyAccessUntil && item.earlyAccessUntil > new Date() && selectedChapters.length > 0"
          :selectable="selectedChapters.length > 0"
          :selected="selectedChapters.some(chapter => chapter.id === item.id)"
          :retry-loading="!!retryChaptersLoading[item.id]"
          :update-progress-loading="!!updateProgressLoading[item.id]"
          @update:selected="toggleSelectChapter(item.id)"
          @update-progress="updateProgress(item.id, $event)"
          @retry="retryChapter(item.id)"
        />
      </template>
    </VVirtualScroll>
  </div>
</template>
