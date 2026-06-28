<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

defineProps<{ filter: Extract<SourceFilterDto, { type: 'checkbox' }> }>()
const model = defineModel<FilterDraftValue>()
const checked = computed({
  get: () => model.value as boolean,
  set: value => (model.value = value),
})
</script>

<template>
  <VCheckbox
    v-model="checked"
    :label="filter.name"
    hide-details
    density="compact"
  />
</template>
