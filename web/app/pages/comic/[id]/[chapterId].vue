<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import type { ReaderSettings } from '~/features/reader/types'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'chapter',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = useRoute('comic-id-chapterId')
const api = createChaptersApi()
const comicsApi = createComicsApi()
const readerSettingsApi = createReaderSettingsApi()
const { t } = useI18n()
const toast = useToast()

const chapter = ref<DetailedChapter>()
const comic = ref<Comic>()
const isLoading = ref(true)

const readerSettings = ref<ReaderSettings>()

async function fetchChapter(): Promise<void> {
  const res = await api.getById(route.params.chapterId)
  if (!res.success) {
    console.error('api.getById', res.error)
    toast.error(res.error.status === 404 || res.error.status === 403
      ? t('sources.asura.chapter.notFound')
      : t('error.unknown'))

    navigateTo({ name: 'comic-id', params: { id: route.params.id } })

    return
  }

  chapter.value = res.data
  syncPolling()
}

async function fetchReaderSettings(): Promise<ReaderSettings[]> {
  const res = await readerSettingsApi.getReaderSettings()
  if (!res.success) {
    console.error('readerSettingsApi.get', res.error)
    toast.error(t('error.unknown'))

    return []
  }

  return res.data
}

async function fetchComic(): Promise<void> {
  const res = await comicsApi.getById(route.params.id)
  if (!res.success) {
    console.error('comicsApi.getById', res.error)
    toast.error(res.error.status === 404 || res.error.status === 403
      ? t('sources.asura.comic.notFound')
      : t('error.unknown'))

    navigateTo({ name: 'library' })

    return
  }

  comic.value = res.data
}

onMounted(async () => {
  const [_, __, settings] = await Promise.all([
    fetchChapter(),
    fetchComic(),
    fetchReaderSettings(),
  ])

  readerSettings.value = settings.find(({ type }) => type === comic.value?.type)
  if (!readerSettings.value) {
    navigateTo({ name: 'comic-id', params: { id: route.params.id } })

    return
  }

  isLoading.value = false
})

const isRetrying = ref(false)

async function retryDownload(): Promise<void> {
  isRetrying.value = true
  const res = await api.retryDownload(route.params.chapterId)
  if (!res.success) {
    console.error('api.retryDownload', res.error)
    toast.error(res.error.status === 404 || res.error.status === 403
      ? t('sources.asura.chapter.notFound')
      : t('error.unknown'))
  }

  isRetrying.value = false
}

watch(() => route.params.chapterId, async (): Promise<void> => {
  isLoading.value = true

  await fetchChapter()

  isLoading.value = false
})

const isInFlight = ref(false)
const { pause, resume } = useIntervalFn(async () => {
  if (isInFlight.value || isLoading.value) {
    return
  }

  isInFlight.value = true
  await fetchChapter()
  isInFlight.value = false
}, 2000, { immediate: false })

function syncPolling(): void {
  if (!chapter.value) {
    pause()

    return
  }

  if (chapter.value.download >= 0 && chapter.value.download < 100) {
    resume()
  } else {
    pause()
  }
}
</script>

<template>
  <div v-if="isLoading" class="d-flex flex-column align-center justify-center w-screen h-screen">
    <VProgressCircular
      indeterminate
      class="mx-auto pa-2"
      color="primary"
      size="100"
    />
  </div>

  <Reader
    v-else-if="comic && chapter && readerSettings"
    :comic
    :chapter
    :settings="readerSettings"
    :retry-download-loading="isRetrying"
    @retry-download="retryDownload"
  />
</template>
