<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../../../comics/types'
import type { BetweenChaptersMode } from '../Card/BetweenChapters.vue'
import type { DetailedChapter } from '~/features/chapters/types'
import type { ReaderSettings } from '~/features/reader/types'

const props = defineProps<{
  comic: Comic
  chapter: DetailedChapter
  settings: ReaderSettings
}>()

const emits = defineEmits<{
  previousChapter: []
  nextChapter: []
  fetchNextChapter: []
  fetchPreviousChapter: []
}>()

const showOverlay = defineModel<boolean>('showOverlay', { required: true })
const page = defineModel<number>('page', { default: 0 })
const images = computed(() => {
  if (props.settings.doublePage) {
    return props.chapter.pageUrls.slice(page.value, page.value + 2)
  }

  return [props.chapter.pageUrls[page.value]]
})

const imageClass = computed(() => {
  const double = props.settings.doublePage

  switch (props.settings.pageScale) {
    case 'fit-width':
      return double ? 'w-50' : 'w-100'
    case 'fit-height':
      return 'h-screen'
    case 'fit-screen':
      return double ? 'h-screen w-50' : 'h-screen w-100'
  }

  return 'h-screen w-100'
})

watch(page, (value) => {
  if (props.settings.doublePage && value % 2 === 1) {
    page.value = value - 1
  }
}, { immediate: true })

const betweenChaptersMode = ref<BetweenChaptersMode>()

watch(() => props.chapter.id, () => {
  betweenChaptersMode.value = undefined
})

watch(page, () => {
  if (betweenChaptersMode.value) {
    betweenChaptersMode.value = undefined
  }
})

watch(betweenChaptersMode, (value) => {
  if (!value) {
    return
  }

  if (value === 'next') {
    emits('fetchNextChapter')
  } else {
    emits('fetchPreviousChapter')
  }
})

function clickLeft(): void {
  if (props.settings.readingMode === 'paged-rtl') {
    nextPage()
  } else {
    previousPage()
  }
}

function clickRight(): void {
  if (props.settings.readingMode === 'paged-rtl') {
    previousPage()
  } else {
    nextPage()
  }
}

function hideOverlay(): void {
  if (showOverlay.value) {
    showOverlay.value = false
  }
}

function previousPage(): void {
  hideOverlay()

  if (page.value > 0) {
    page.value = page.value - (props.settings.doublePage ? 2 : 1)
  } else if (!betweenChaptersMode.value) {
    betweenChaptersMode.value = 'previous'
  } else {
    emits('previousChapter')
  }
}

function nextPage(): void {
  hideOverlay()

  if (page.value < props.chapter.pageUrls.length - (props.settings.doublePage ? 2 : 1)) {
    page.value = page.value + (props.settings.doublePage ? 2 : 1)
  } else if (!betweenChaptersMode.value) {
    betweenChaptersMode.value = 'next'
  } else {
    emits('nextChapter')
  }
}
</script>

<template>
  <div class="h-screen w-screen overflow-auto" @click="showOverlay = !showOverlay">
    <div class="position-relative" style="min-height: 100%;">
      <div
        class="position-absolute top-0 bottom-0 left-0 w-33"
        style="z-index: 1;"
        @click.stop="clickLeft"
      />

      <div
        class="position-absolute top-0 bottom-0 right-0 w-33"
        style="z-index: 1;"
        @click.stop="clickRight"
      />

      <ReaderCardBetweenChapters
        v-if="betweenChaptersMode"
        :comic-id="comic.id"
        :current-chapter="chapter"
        :next-chapter="betweenChaptersMode === 'next' ? chapter.next : chapter.previous"
        :mode="betweenChaptersMode"
      />

      <div
        v-else
        class="d-flex justify-center align-center min-h-screen"
        :class="{
          'flex-row-reverse': settings.readingMode === 'paged-rtl',
          'overflow-auto': settings.pageScale !== 'fit-screen',
        }"
      >
        <VImg
          v-for="src in images"
          :key="src"
          :src="src"
          eager
          :class="imageClass"
        />
      </div>
    </div>
  </div>
</template>
