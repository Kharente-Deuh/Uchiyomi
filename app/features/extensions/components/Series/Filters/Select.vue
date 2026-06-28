<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

const props = defineProps<{ filter: Extract<SourceFilterDto, { type: 'select' }> }>()
const model = defineModel<FilterDraftValue>()
const selected = computed({
  get: () => model.value as number,
  set: value => (model.value = value),
})
const items = computed(() => props.filter.values.map((title, value) => ({ title, value })))
</script>

<template>
  <VSelect
    v-model="selected"
    :items="items"
    :label="filter.name"
    hide-details
    density="compact"
  />
</template>
