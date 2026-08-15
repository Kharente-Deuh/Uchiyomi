<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const props = defineProps<{
  slug: string
  comicOriginUrl: string
}>()

const { sort, chapters, loading, fetchChapters } = useAsuraChapters()

watch(() => props.slug, () => {
  fetchChapters()
}, { immediate: true })
</script>

<template>
  <div class="d-flex flex-column w-100 position-relative bg-surface" style="border-radius: 12px; max-height: 40rem;">
    <div class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center" style="z-index: 1; border-top-left-radius: 12px; border-top-right-radius: 12px;">
      <span class="text-title-large font-weight-bold">{{ $t('sources.asurascans.comic.chaptersCount', { count: chapters.length }) }}</span>
      <VBtn
        variant="tonal"
        class="text-body-medium"
        color="surfaceVariant"
        :text="sort === 'asc' ? $t('common.sort.oldest') : $t('common.sort.latest')"
        :prepend-icon="sort === 'desc' ? 'fa6-solid:arrow-down-short-wide' : 'fa6-solid:arrow-up-short-wide'"
        @click="sort = sort === 'asc' ? 'desc' : 'asc'"
      />
    </div>

    <div v-if="loading || chapters.length === 0" class="d-flex justify-center align-center w-100 pa-4">
      <VProgressCircular
        v-if="loading"
        class="m-auto"
        indeterminate
        color="primary"
      />
      <span v-else class="text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: 0 }) }}</span>
    </div>
    <VVirtualScroll :items="chapters" style="border-bottom-left-radius: 12px; border-bottom-right-radius: 12px;">
      <template #default="{ item }">
        <AsuraComicChaptersItem
          :chapter="item"
          :comic-origin-url="props.comicOriginUrl"
        />
      </template>
    </VVirtualScroll>
  </div>
</template>
