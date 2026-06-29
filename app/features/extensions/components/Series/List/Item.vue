<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { MangaChapterSummaryDto } from '~~/shared/dto/catalogue/manga-chapter-summary.dto'
import type { SourceSearchItemDto } from '#shared/dto/catalogue/source-search.dto'

defineProps<{
  manga: SourceSearchItemDto
  status?: ChapterSummaryStatus
  summary?: MangaChapterSummaryDto
}>()

const relativeTime = useRelativeTime()
</script>

<template>
  <div class="d-flex ga-4 align-center text-truncate w-100 series-list-item" :class="manga.inLibrary ? 'series-list-item-library' : 'series-list-item-not-library'">
    <VImg
      v-if="manga.thumbnailUrl"
      :src="manga.thumbnailUrl"
      aspect-ratio="9/16"
      cover
      width="96px"
      height="147px"
      rounded="lg"
      class="border-thin"
    />

    <div class="d-flex flex-column text-truncate w-100 h-100 py-1 ga-1 justify-space-between">
      <div class="d-flex align-center justify-space-between ga-2 text-truncate">
        <span class="text-title-medium font-weight-bold text-truncate">{{ manga.title }}</span>
        <AtomLink
          v-if="manga.sourceUrl"
          external
          new-tab
          :to="manga.sourceUrl"
        >
          <VBtn
            icon="fa6-solid:arrow-up-right-from-square"
            size="x-small"
            class="border-thin-secondary"
            color="secondary"
          />
        </AtomLink>
      </div>
      <template v-if="status === 'success' && summary">
        <span class="text-body-medium opacity-70 text-truncate">{{ $t('series.summary.chapterCount', { count: summary.chapterCount }) }}</span>
        <template v-if="summary.lastChapter">
          <div class="d-flex align-center justify-space-between ga-2 text-truncate">
            <span class="text-body-medium opacity-70 text-truncate">{{ summary.lastChapter.name }}</span>
            <span class="text-body-medium opacity-70">{{ relativeTime(summary.lastChapter.uploadedAt) }}</span>
          </div>
        </template>
      </template>
      <template v-else-if="status === 'loading' || status === 'queued'">
        <VSkeletonLoader
          type="text"
          width="50%"
          class="mb-1"
        />
        <VSkeletonLoader
          type="text"
          class="ma-0"
          width="100%"
        />
      </template>
      <template v-else-if="status === 'error'">
        <span class="text-body-small text-medium-emphasis text-error ma-auto">{{ $t('series.summary.errors.loadFailed') }}</span>
      </template>

      <div class="d-flex align-center justify-space-between ga-4 text-truncate mt-auto">
        <VBtn
          v-if="!manga.inLibrary"
          class="ml-auto border-thin-primary"
          icon="fa6-solid:plus"
          color="primary"
          size="x-small"
        />
        <VChip
          v-else
          prepend-icon="fa6-solid:check"

          class="ml-auto"
          color="secondary"
          density="compact"
          rounded="pill"
          :text="$t('series.summary.alreadyInLibrary')"
        />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.series-list-item {
  transition: all 0.3s ease-in-out;

  &-library {
    :deep(img) {
      filter: brightness(30%) !important;
    }
  }
}
</style>
