<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
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
} = useSearchSeriesSourcesComposable()

const searchTypeItems = computed(() => {
  const currentSource = sourcesItems.value.find(source => source.value === selectedSourceId.value)

  return [
    { title: $t('sources.series.searchType.popular'), value: 'popular' },
    ...(currentSource?.supportsLatest ? [{ title: $t('sources.series.searchType.latest'), value: 'latest' }] : []),
    { title: $t('sources.series.searchType.search'), value: 'search' },
  ]
})

const { mobile } = useDisplay()
</script>

<template>
  <div class="d-flex flex-column ga-6 mt-2">
    <div class="d-flex align-center ga-8 justify-space-between w-100">
      <div class="d-flex ga-5 align-center transition-smooth justify-space-between w-100">
        <div class="d-flex ga-3 align-center">
          <span class="text-title-large font-weight-bold font-title">{{ $t('series.title') }}</span>
        </div>
        <AtomInputSearch
          v-if="searchType === 'search'"
          v-model="searchFilter"
          style="max-width: 30rem;"
        />
        <div v-if="sourcesItems.length > 0" class="d-flex ga-3 align-center">
          <MoleculeBtnList v-model="searchType" :items="searchTypeItems" />
          <MoleculeBtnList
            v-show="sourcesItems.length > 1"
            :items="sourcesItems"
            :model-value="selectedSourceId ?? ''"
            @update:model-value="selectedSourceId = $event"
          />
        </div>
      </div>
    </div>

    <VProgressLinear
      v-if="searchLoading"
      indeterminate
      class="w-100 rounded-lg"
    />

    <div class="series-list-grid">
      <ExtensionsSeriesListItem
        v-for="manga in series"
        :key="manga.id"
        :manga
        :status="chapterStatusOf(manga.id)"
        :summary="chapterSummaryOf(manga.id)"
      />
    </div>
    <MoleculePaginationFooter
      v-if="!mobile"
      v-model="page"
      :has-next-page
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
