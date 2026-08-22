<script setup lang="ts">
import type { AdjacentChapter, Chapter } from '~/features/chapters/types'

export type BetweenChaptersMode = 'next' | 'previous'

defineProps<{
  comicId: string
  currentChapter: Chapter
  nextChapter?: AdjacentChapter
  mode: BetweenChaptersMode
}>()
</script>

<template>
  <div class="d-flex flex-column ga-8 justify-center h-screen mx-auto" style="max-width:20rem;">
    <div class="d-flex flex-column ga-2 text-truncate">
      <span class="text-body-medium">{{ $t('comic.betweenChapters.current') }} :</span>
      <span class="text-title-large font-weight-bold"> {{ $t('feed.chapter.number', { number: currentChapter.number }) }} </span>
      <span class="text-title-medium text-medium-emphasis"> {{ currentChapter.title }} </span>
    </div>

    <VDivider />

    <div v-if="nextChapter" class="d-flex flex-column ga-2 text-truncate">
      <span class="text-body-medium">{{ mode === 'next' ? $t('comic.betweenChapters.next') : $t('comic.betweenChapters.previous') }} :</span>
      <span class="text-title-large font-weight-bold"> {{ $t('feed.chapter.number', { number: nextChapter.number }) }} </span>
      <span class="text-title-medium text-medium-emphasis"> {{ nextChapter.title }} </span>
    </div>
    <div v-else class="d-flex flex-column ga-4 align-center justify-center">
      <span class="text-body-medium">{{ mode === 'next' ? $t('comic.betweenChapters.endOfComic') : $t('comic.betweenChapters.startOfComic') }}</span>

      <AtomLink :to="`/comic/${comicId}`">
        <VBtn
          :text="$t('comic.chapter.exitToComic')"
          color="secondary"
          variant="flat"
        />
      </AtomLink>
    </div>
  </div>
</template>
