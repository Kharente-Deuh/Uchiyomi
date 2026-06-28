<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

defineProps<{ filter: Exclude<SourceFilterDto, { type: 'group' }> }>()
const model = defineModel<FilterDraftValue>()
</script>

<template>
  <ExtensionsSeriesFiltersCheckBox
    v-if="filter.type === 'checkbox'"
    v-model="model"
    :filter="filter"
  />
  <ExtensionsSeriesFiltersTriState
    v-else-if="filter.type === 'tristate'"
    v-model="model"
    :filter="filter"
  />
  <ExtensionsSeriesFiltersSelect
    v-else-if="filter.type === 'select'"
    v-model="model"
    :filter="filter"
  />
  <ExtensionsSeriesFiltersText
    v-else-if="filter.type === 'text'"
    v-model="model"
    :filter="filter"
  />
  <ExtensionsSeriesFiltersSort
    v-else-if="filter.type === 'sort'"
    v-model="model"
    :filter="filter"
  />
  <div
    v-else-if="filter.type === 'header'"
    class="text-body-medium font-weight-bold mt-2"
  >
    {{ filter.name }}
  </div>
  <VDivider v-else-if="filter.type === 'separator'" />
</template>
