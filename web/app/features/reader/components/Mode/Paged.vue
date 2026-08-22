<script setup lang="ts">
import type { Comic } from '../../../comics/types'
import type { BetweenChaptersMode } from '../Card/BetweenChapters.vue'
import type { DetailedChapter } from '~/features/chapters/types'
import type { ReaderSettings } from '~/features/reader/types'

const props = defineProps<{
  comic: Comic
  chapter: DetailedChapter
  settings: ReaderSettings
  startPage?: number
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
  page.value = 0
  betweenChaptersMode.value = undefined
})

watch(page, () => {
  if (betweenChaptersMode.value) {
    betweenChaptersMode.value = undefined
  }

  if (showOverlay.value) {
    showOverlay.value = false
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

function previousPage(): void {
  if (page.value > 0) {
    page.value = page.value - (props.settings.doublePage ? 2 : 1)
  } else if (!betweenChaptersMode.value) {
    betweenChaptersMode.value = 'previous'
  } else {
    emits('previousChapter')
  }
}

function nextPage(): void {
  if (page.value < props.chapter.pageUrls.length - 1) {
    page.value = page.value + (props.settings.doublePage ? 2 : 1)
  } else if (!betweenChaptersMode.value) {
    betweenChaptersMode.value = 'next'
  } else {
    emits('nextChapter')
  }
}
</script>

<template>
  <div class="h-screen w-screen overflow-auto position-relative" @click="showOverlay = !showOverlay">
    <div
      class="position-fixed top-0 left-0 h-screen w-33"
      style="z-index: 1;"
      @click.stop="clickLeft"
    />

    <div
      class="position-fixed top-0 right-0 h-screen w-33"
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
</template>
