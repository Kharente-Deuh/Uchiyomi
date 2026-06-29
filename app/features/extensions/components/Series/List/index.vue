<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceSearchQueryType } from '#shared/dto/catalogue/source-search.dto'

const {
  searchType,
  selectedSourceId,
  sourcesItems,
  series,
  chapterStatusOf,
  chapterSummaryOf,
  page,
  hasNextPage,
  searchFilter,
  searchLoading,
  definitions,
  filtersDraft,
  filtersActiveCount,
  filtersLoading,
  applyFilters,
  resetFilters,
} = useSearchSeriesSourcesComposable()

const showFilters = ref(false)

const searchTypeItems = computed((): { title: string, value: SourceSearchQueryType }[] => {
  const currentSource = sourcesItems.value.find(source => source.value === selectedSourceId.value)
  const items: { title: string, value: SourceSearchQueryType }[] = [
    { title: $t('sources.series.searchType.popular'), value: 'popular' },
  ]

  if (currentSource?.supportsLatest) {
    items.push({ title: $t('sources.series.searchType.latest'), value: 'latest' })
  }

  items.push({ title: $t('sources.series.searchType.search'), value: 'search' })

  return items
})

const { mobile } = useDisplay()

const loadMoreSentinel = useTemplateRef<HTMLElement>('loadMoreSentinel')
useIntersectionObserver(loadMoreSentinel, ([entry]) => {
  if (entry?.isIntersecting && mobile.value && !searchLoading.value && hasNextPage.value) {
    page.value += 1
  }
})
</script>

<template>
  <div class="d-flex flex-column ga-6 mt-2 position-relative">
    <ExtensionsSeriesListMobileHeader
      v-if="mobile"
      v-model:search-filter="searchFilter"
      v-model:search-type="searchType"
      v-model:selected-source-id="selectedSourceId"
      :search-type-items="searchTypeItems"
      :fetch-filters-loading="filtersLoading"
      :sources-items="sourcesItems"
      :filters-active-count="filtersActiveCount"
      @open-filters="showFilters = true"
    />
    <ExtensionsSeriesListDesktopHeader
      v-else
      v-model:search-filter="searchFilter"
      v-model:search-type="searchType"
      v-model:selected-source-id="selectedSourceId"
      :search-type-items="searchTypeItems"
      :sources-items="sourcesItems"
      :fetch-filters-loading="filtersLoading"
      :filters-active-count="filtersActiveCount"
      @open-filters="showFilters = true"
    />
    <VProgressLinear
      v-if="searchLoading"
      indeterminate
      class="w-100 rounded-lg"
    />

    <div
      v-if="series.length > 0"
      class="series-list-grid"
      :class="{ 'px-4': mobile }"
    >
      <ExtensionsSeriesListItem
        v-for="manga in series"
        :key="manga.id"
        :manga
        :status="chapterStatusOf(manga.id)"
        :summary="chapterSummaryOf(manga.id)"
      />
    </div>
    <span v-else class="text-body-medium text-medium-emphasis w-100 text-center">{{ $t('sources.series.search.empty') }}</span>
    <MoleculePaginationFooter
      v-if="!mobile && (hasNextPage || page > 1)"
      v-model="page"
      :has-next-page
    />

    <div
      v-if="mobile"
      ref="loadMoreSentinel"
      class="d-flex justify-center py-4"
    >
      <VProgressCircular
        v-if="searchLoading && page > 1"
        color="secondary"
        indeterminate
      />
    </div>

    <ExtensionsSeriesFiltersModal
      v-model="showFilters"
      v-model:draft="filtersDraft"
      :definitions="definitions"
      :loading="filtersLoading"
      @apply="applyFilters"
      @reset="resetFilters"
    />
  </div>
</template>

<style lang="scss">
.series-list-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.5rem;
}

@media screen and (max-width: 1300px) {
  .series-list-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media screen and (max-width: 790px) {
  .series-list-grid {
    grid-template-columns: 1fr;
  }
}
</style>
