<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'
import { VVirtualScroll } from 'vuetify/components'
import { pageFromVisibleIndices, recordIntersection } from './visible-page'

defineProps<{
  comic: Comic
  chapter: DetailedChapter
}>()

const page = defineModel<number>('page', { default: 0 })
const showOverlay = defineModel<boolean>('showOverlay', { required: true })

const virtualScrollRef = ref<InstanceType<typeof VVirtualScroll>>()
const scrollRoot = computed(() => {
  const el = virtualScrollRef.value?.$el

  return el instanceof HTMLElement ? el : undefined
})

const internalPage = ref<number>()
const movingToPage = ref<number | null>(page.value === 0 ? null : page.value)

const unwatchVirtualScroll = watch(virtualScrollRef, async (value) => {
  if (!value) {
    return
  }

  if (movingToPage.value !== null) {
    await nextTick()
    value.scrollToIndex(movingToPage.value)
  }

  unwatchVirtualScroll()
})

watch(page, (value) => {
  if (value === internalPage.value) {
    return
  }

  movingToPage.value = value
  virtualScrollRef.value?.scrollToIndex(value)
})

const intersectingPages = new Set<number>()

function onIntersecting(index: number, isIntersecting: boolean): void {
  recordIntersection(intersectingPages, index, isIntersecting)

  const current = pageFromVisibleIndices(intersectingPages, {
    movingToPage: movingToPage.value,
  })
  if (current === undefined) {
    return
  }

  internalPage.value = current

  if (movingToPage.value !== null) {
    return
  }

  if (page.value !== current) {
    page.value = current
  }
}

function onScrollEnd(): void {
  if (movingToPage.value === null) {
    return
  }

  intersectingPages.clear()
  internalPage.value = movingToPage.value
  movingToPage.value = null
}

const preventFirstScroll = ref(true)

function onScroll(event: Event): void {
  if (preventFirstScroll.value) {
    preventFirstScroll.value = false

    return
  }

  if (movingToPage.value !== null) {
    return
  }

  const el = event.currentTarget as HTMLElement
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight <= 1
  if (!atBottom && showOverlay.value) {
    showOverlay.value = false
  } else if (atBottom && !showOverlay.value) {
    showOverlay.value = true
  }
}
</script>

<template>
  <div
    class="h-screen w-screen mx-auto"
    @click="showOverlay = !showOverlay"
  >
    <VVirtualScroll
      ref="virtualScrollRef"
      class="h-100"
      :items="chapter.pageUrls"
      @scroll="onScroll"
      @scrollend="onScrollEnd"
    >
      <template #default="{ item, index }">
        <ReaderModeScrollImg
          :src="item"
          style="max-width: 70rem; margin-right: auto; margin-left: auto;"
          :root="scrollRoot"
          @intersecting="onIntersecting(index, $event)"
        />
      </template>
    </VVirtualScroll>
  </div>
</template>
