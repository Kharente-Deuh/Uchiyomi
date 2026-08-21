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
const { smAndDown } = useDisplay()

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

  <template v-else-if="chapter">
    <div v-if="chapter.download !== 100" class="d-flex flex-column w-screen h-screen justify-center align-center ga-4">
      <VProgressCircular
        v-if="chapter.download !== -1"
        :indeterminate="chapter.download === 0"
        :value="chapter.download"
        size="48"
        color="primary"
      />
      <template v-if="chapter.download === -1">
        <VIcon
          icon="fa6-solid:circle-exclamation"
          size="48"
          color="error"
        />
        <span class="text-body-large text-medium-emphasis">{{ $t('comic.chapter.download.error') }}</span>
        <VBtn
          color="error"
          class="border-thin-error"
          :text="$t('comic.chapter.download.retry')"
          :loading="isRetrying"
          @click="retryDownload"
        />
      </template>

      <div class="d-flex align-center flex-wrap ga-4">
        <AtomLink v-if="chapter.previousChapterId" :to="{ name: 'comic-id-chapterId', params: { id: route.params.id, chapterId: chapter.previousChapterId } }">
          <VBtn
            v-if="chapter.previousChapterId"
            color="primary"
            :class="{
              'w-100': !chapter.previousChapterId,
              'w-auto': chapter.previousChapterId,
              'px-3': smAndDown,
            }"
            :prepend-icon="smAndDown ? undefined : 'fa6-solid:angle-left'"
            :text="smAndDown ? undefined : $t('comic.chapter.previous')"
            :icon="smAndDown ? 'fa6-solid:angle-left' : undefined"
            class="border-thin-primary"
          />
        </AtomLink>

        <AtomLink v-if="chapter.nextChapterId" :to="{ name: 'comic-id-chapterId', params: { id: route.params.id, chapterId: chapter.nextChapterId } }">
          <VBtn
            v-if="chapter.nextChapterId"
            color="primary"
            :icon="smAndDown ? 'fa6-solid:angle-right' : undefined"
            :text="smAndDown ? undefined : $t('comic.chapter.next')"
            :append-icon="smAndDown ? undefined : 'fa6-solid:angle-right'"
            :class="{
              'w-100': !chapter.nextChapterId,
              'w-auto': chapter.nextChapterId,
              'px-3': smAndDown,
            }"
            class="border-thin-primary"
          />
        </AtomLink>
      </div>
      <AtomLink :to="{ name: 'comic-id', params: { id: route.params.id } }">
        <VBtn
          class="border-thin-secondary"
          :class="{
            'w-100': smAndDown,
            'w-fit-content': !smAndDown,
          }"
          :text="$t('comic.chapter.exitToComic')"
          color="secondary"
        />
      </AtomLink>
    </div>
  </template>
</template>
