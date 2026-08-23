<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'

defineProps<{
  comic: Comic
  chapter: DetailedChapter
}>()
</script>

<template>
  <div
    class="d-flex align-center w-100 bg-surface border-t-thin pa-4 position-fixed bottom-0 left-0 right-0"
    style="z-index: 2;"
  >
    <div class="d-flex ga-4 justify-space-between align-center w-100 mx-auto" style="max-width: 80rem;">
      <AtomLink :to="chapter.previous ? `/comic/${comic.id}/${chapter.previous.id}?ignoreProgress=true` : undefined">
        <VBtn
          :disabled="!chapter.previous"
          :variant="chapter.previous ? 'tonal' : 'text'"
          prepend-icon="fa6-solid:angle-left"
          :text="$t('comic.chapter.previous')"
          :class="{ 'border-thin-primary': chapter.previous }"
        />
      </AtomLink>

      <ReaderOverlayFooterChaptersMenu :comic-id="comic.id" :current-chapter="chapter" />

      <AtomLink :to="chapter.next ? `/comic/${comic.id}/${chapter.next.id}?ignoreProgress=true` : undefined">
        <VBtn
          :disabled="!chapter.next"
          :variant="chapter.next ? 'tonal' : 'text'"
          append-icon="fa6-solid:angle-right"
          :text="$t('comic.chapter.next')"
          :class="{ 'border-thin-primary': chapter.next }"
        />
      </AtomLink>
    </div>
  </div>
</template>

<style lang="scss">
.chapter-list-item {
  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}
</style>
