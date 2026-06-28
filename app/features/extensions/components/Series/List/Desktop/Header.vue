<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceSearchQueryType } from '~~/shared/dto/catalogue/source-search.dto'

defineProps<{
  searchTypeItems: { title: string, value: SourceSearchQueryType }[]
  sourcesItems: { value: string, title: string, supportsLatest: boolean }[]
  filtersActiveCount: number
  fetchFiltersLoading: boolean
}>()
const emit = defineEmits<{ openFilters: [] }>()

const searchType = defineModel<SourceSearchQueryType>('searchType', { required: true })
const searchFilter = defineModel<string | undefined>('searchFilter', { required: true })
const selectedSourceId = defineModel<string | undefined>('selectedSourceId', { required: true })
</script>

<template>
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
        <VBadge
          v-if="searchType === 'search'"
          :model-value="filtersActiveCount > 0"
          :content="filtersActiveCount"
          color="primary"
        >
          <VChip
            color="secondary"
            class="border-thin-secondary cursor-pointer"
            rounded="pill"
            @click="emit('openFilters')"
          >
            <VProgressCircular
              v-if="fetchFiltersLoading"
              color="secondary"
              indeterminate
              width="2"
              size="16"
            />
            <VIcon v-else icon="fa6-solid:filter" />
          </VChip>
        </VBadge>
      </div>
    </div>
  </div>
</template>
