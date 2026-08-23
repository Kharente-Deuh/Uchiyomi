<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Chapter } from '~/features/chapters/types'
import { VVirtualScroll } from 'vuetify/components'

const props = defineProps<{
  comicId: string
  currentChapter: Chapter
}>()

const chaptersApi = createChaptersApi()

const chapters = ref<Chapter[]>([])
const chaptersLoading = ref(false)

watch(() => props.comicId, () => {
  fetchChapters()
}, { immediate: true })

async function fetchChapters(): Promise<void> {
  chaptersLoading.value = true
  const res = await chaptersApi.getByComicId(props.comicId)

  if (res.success) {
    chapters.value = res.data
  } else {
    console.error('chaptersApi.getByComicId', res.error)
  }

  chaptersLoading.value = false
}

const showMenu = ref(false)
const { pause, resume } = useIntervalFn(() => {
  void pollChapters()
}, 2000, { immediate: false })

function syncPolling(): void {
  const hasChaptersToPoll = chapters.value.some(chapter => chapter.download >= 0 && chapter.download < 100)

  if (showMenu.value && hasChaptersToPoll) {
    resume()
  } else {
    pause()
  }
}

watch(showMenu, (value: boolean) => {
  if (value) {
    syncPolling()
  }
})

const virtualScrollRef = ref<InstanceType<typeof VVirtualScroll>>()
watch(virtualScrollRef, (value) => {
  if (!value) {
    return
  }

  const currentChapterIndex = chapters.value.findIndex(chapter => chapter.id === props.currentChapter.id)
  if (currentChapterIndex === -1) {
    return
  }

  value.scrollToIndex(currentChapterIndex)
})

const isInFlight = ref(false)
async function pollChapters(): Promise<void> {
  if (isInFlight.value || chaptersLoading.value) {
    return
  }

  isInFlight.value = true

  const chaptersIds: string[] = []
  for (const chapter of chapters.value) {
    if (chapter.download >= 0 && chapter.download < 100) {
      chaptersIds.push(chapter.id)
    }
  }

  if (chaptersIds.length === 0) {
    isInFlight.value = false

    return
  }

  const res = await chaptersApi.getByIds(chaptersIds)
  if (!res.success) {
    console.error('chaptersApi.getByIds', res.error)
    isInFlight.value = false

    return
  }

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

  isInFlight.value = false
  syncPolling()
}
</script>

<template>
  <VBtn
    prepend-icon="fa6-solid:bars"
    color="grey"
    :loading="chaptersLoading"
    :disabled="!chaptersLoading && chapters.length === 0"
    :append-icon="showMenu ? 'fa6-solid:chevron-up' : 'fa6-solid:chevron-down'"
  >
    {{ $t('feed.chapter.number', { number: currentChapter.number }) }}
    <VMenu
      v-model="showMenu"
      activator="parent"
      location="top center"
      offset="16"
    >
      <VVirtualScroll
        ref="virtualScrollRef"
        :items="chapters"
        style="max-width: 15rem; max-height: 20rem; border-radius: 12px;"
        class="border-thin pa-2 bg-surface"
      >
        <template #default="{ item }">
          <AtomLink :to="`/comic/${comicId}/${item.id}`">
            <div class="pa-2 w-100 rounded-lg reader-chapters-menu-item" :class="{ 'reader-chapters-menu-item--selected': item.id === currentChapter.id }">
              <div class="d-flex ga-4 justify-space-between align-center text-truncate">
                <div class="d-flex flex-column">
                  <span class="text-body-large font-weight-bold">{{ $t('feed.chapter.number', { number: item.number }) }}</span>
                  <span class="text-body-medium text-medium-emphasis">{{ item.title }}</span>
                </div>
                <VIcon
                  v-if="item.download === -1"
                  color="error"
                  class="ml-2"
                  icon="fa6-solid:exclamation"
                />

                <VProgressCircular
                  v-else-if="item.download >= 0 && item.download < 100"
                  class="ml-2"
                  :model-value="item.download"
                  size="18"
                  :indeterminate="item.download === 0"
                  width="2"
                  color="primary"
                />
              </div>
            </div>
          </AtomLink>
        </template>
      </vvirtualscroll>
    </VMenu>
  </VBtn>
</template>

<style lang="scss">
.reader-chapters-menu-item {
  &:hover {
    color: rgb(var(--v-theme-primary));
  }

  &--selected {
    background-color: rgba(var(--v-theme-primary), 0.1);
    color: rgb(var(--v-theme-primary));
  }
}
</style>
