<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceSearchQueryType } from '~~/shared/dto/catalogue/source-search.dto'

defineProps<{
  searchTypeItems: { title: string, value: SourceSearchQueryType }[]
  sourcesItems: { value: string, title: string, supportsLatest: boolean }[]
}>()

const searchType = defineModel<SourceSearchQueryType>('searchType', { required: true })
const searchFilter = defineModel<string | undefined>('searchFilter', { required: true })
const selectedSourceId = defineModel<string | undefined>('selectedSourceId', { required: true })
</script>

<template>
  <div class="series-list-mobile-header d-flex flex-column ga-2 w-100 pb-4 border-b-thin bg-background">
    <div class="d-flex align-center justify-space-between ga-4 px-6 py-1">
      <span class=" text-medium-emphasis text-body-small text-uppercase py-1" style="letter-spacing: 0.08em;">
        {{ $t('series.title') }}
      </span>

      <div v-if="sourcesItems.length > 0" class="d-flex align-center ga-3">
        <MoleculeBtnList v-model="searchType" :items="searchTypeItems" />
        <MoleculeBtnList
          v-show="sourcesItems.length > 1"
          :items="sourcesItems"
          :model-value="selectedSourceId ?? ''"
          @update:model-value="selectedSourceId = $event"
        />
      </div>
    </div>
    <AtomInputSearch
      v-if="searchType === 'search'"
      v-model="searchFilter"
      class="px-4"
    />
  </div>
</template>

<style lang="scss" scoped>
.series-list-mobile-header {
  position: sticky;
  top: calc(var(--page-header-height, 0px) + var(--tabs-height, 0px));
  z-index: 5;
}
</style>
