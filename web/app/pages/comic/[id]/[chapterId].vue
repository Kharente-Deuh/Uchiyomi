<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'chapter',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})

const route = computed(() => useRoute('comic-id-chapterId'))
const router = useRouter()

const {
  page,
  chapter,
  comic,
  readerSettings,
  isLoading,
  fetchNextChapter,
  fetchPreviousChapter,
  retryDownload,
  retryDownloadLoading,
  nextChapter,
  previousChapter,
  startEnd,
} = useReader({
  comicId: route.value.params.id,
  chapterId: computed(() => route.value.params.chapterId),
  ignoreProgress: computed(() => route.value.query.ignoreProgress === 'true'),
})

function toPreviousChapter(): void {
  if (!previousChapter.value) {
    return
  }

  startEnd.value = true

  updateRoute(previousChapter.value.id)
}

function toNextChapter(): void {
  if (!nextChapter.value) {
    return
  }

  updateRoute(nextChapter.value.id)
}

function updateRoute(chapterId: string): void {
  router.replace({
    name: 'comic-id-chapterId',
    params: {
      id: route.value.params.id,
      chapterId,
    },
  })
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
    v-model:page="page"
    :comic
    :chapter
    :settings="readerSettings"
    :retry-download-loading
    @retry-download="retryDownload"
    @fetch-next-chapter="fetchNextChapter"
    @fetch-previous-chapter="fetchPreviousChapter"
    @previous-chapter="toPreviousChapter"
    @next-chapter="toNextChapter"
  />
</template>
