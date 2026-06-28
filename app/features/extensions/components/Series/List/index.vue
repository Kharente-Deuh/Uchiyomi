<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
const { searchType, selectedSourceId, sourcesItems } = useSearchSeriesSourcesComposable()
</script>

<template>
  <div class="d-flex flex-column ga-6">
    <div class="d-flex align-center ga-8 justify-space-between w-100">
      <div class="d-flex ga-3 align-center transition-smooth justify-space-between w-100">
        <div class="d-flex ga-3 align-center">
          <span class="text-title-large font-weight-bold font-title">{{ $t('series.title') }}</span>
        </div>
        <div class="d-flex ga-3 align-center">
          <AtomBtnToggle
            v-for="type in ['popular', 'latest', 'search']"
            :key="type"
            :text="$t(`sources.series.searchType.${type}`)"
            :model-value="searchType === type"
            type="on-off"
            @update:model-value="searchType = type"
          />

          <VSelect
            v-model="selectedSourceId"
            density="compact"
            :items="sourcesItems"
            label="Source"
            hide-details
          />
        </div>
      </div>
    </div>

    <div class="d-flex ga-4 overflow-x-auto custom-scrollbar pb-4" />
  </div>
</template>
