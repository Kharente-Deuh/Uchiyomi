<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

defineProps<{ filter: Extract<SourceFilterDto, { type: 'text' }> }>()
const model = defineModel<FilterDraftValue>()
const text = computed({
  get: () => model.value as string,
  set: value => (model.value = value),
})
</script>

<template>
  <VTextField
    v-model="text"
    :label="filter.name"
    hide-details
    density="compact"
  />
</template>
