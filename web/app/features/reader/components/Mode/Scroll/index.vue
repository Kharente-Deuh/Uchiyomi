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

const restoredPage = page.value
const restoring = ref(restoredPage > 0)

const unwatchVirtualScroll = watch(virtualScrollRef, async (value) => {
  if (!value) {
    return
  }

  if (restoring.value) {
    await nextTick()
    value.scrollToIndex(restoredPage)
  }

  unwatchVirtualScroll()
})

const intersectingPages = new Set<number>()

function onIntersecting(index: number, isIntersecting: boolean): void {
  recordIntersection(intersectingPages, index, isIntersecting)

  const current = pageFromVisibleIndices(intersectingPages, {
    restoredPage,
    restoring: restoring.value,
  })
  if (current === undefined) {
    return
  }

  if (restoring.value) {
    restoring.value = false
    for (const intersectingIndex of intersectingPages) {
      if (intersectingIndex < restoredPage) {
        intersectingPages.delete(intersectingIndex)
      }
    }
  }

  if (page.value !== current) {
    page.value = current
  }
}

const preventFirstScroll = ref(true)

function onScroll(event: Event): void {
  if (preventFirstScroll.value) {
    preventFirstScroll.value = false

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
  <div class="h-screen w-screen overflow-hidden" @click="showOverlay = !showOverlay">
    <VVirtualScroll
      ref="virtualScrollRef"
      class="h-100"
      :items="chapter.pageUrls"
      @scroll="onScroll"
    >
      <template #default="{ item, index }">
        <ReaderModeScrollImg
          :src="item"
          :root="scrollRoot"
          @intersecting="onIntersecting(index, $event)"
        />
      </template>
    </VVirtualScroll>
  </div>
</template>
