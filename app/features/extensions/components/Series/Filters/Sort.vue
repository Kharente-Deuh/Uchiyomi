<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SortStateDto, SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

const props = defineProps<{ filter: Extract<SourceFilterDto, { type: 'sort' }> }>()
const model = defineModel<FilterDraftValue>()
const state = computed(() => (model.value as SortStateDto | undefined) ?? { ascending: false, index: 0 })
const items = computed(() => props.filter.values.map((title, index) => ({ title, value: index })))

function setIndex(index: number): void {
  model.value = { ascending: state.value.ascending, index }
}

function toggleDirection(): void {
  model.value = { ascending: !state.value.ascending, index: state.value.index }
}
</script>

<template>
  <div class="d-flex align-center ga-2">
    <VSelect
      :model-value="state.index"
      :items="items"
      :label="filter.name"
      hide-details
      density="compact"
      @update:model-value="setIndex"
    />
    <VBtn
      :icon="state.ascending ? 'fa6-solid:arrow-up-short-wide' : 'fa6-solid:arrow-down-wide-short'"
      size="small"
      variant="text"
      @click="toggleDirection"
    />
  </div>
</template>
