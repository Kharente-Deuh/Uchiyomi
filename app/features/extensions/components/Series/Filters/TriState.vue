<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto, TriStateValue } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'

defineProps<{ filter: Extract<SourceFilterDto, { type: 'tristate' }> }>()
const model = defineModel<FilterDraftValue>()

const ORDER: TriStateValue[] = ['IGNORE', 'INCLUDE', 'EXCLUDE']
const ICON: Record<TriStateValue, string> = {
  IGNORE: 'fa6-regular:square',
  INCLUDE: 'fa6-solid:square-check',
  EXCLUDE: 'fa6-solid:square-xmark',
}
const state = computed(() => model.value as TriStateValue)

function cycle(): void {
  model.value = ORDER[(ORDER.indexOf(state.value) + 1) % ORDER.length] as TriStateValue
}
</script>

<template>
  <div
    class="d-flex align-center ga-3 cursor-pointer py-1"
    @click="cycle"
  >
    <VIcon
      :icon="ICON[state]"
      size="small"
      :color="state === 'INCLUDE' ? 'primary' : state === 'EXCLUDE' ? 'error' : undefined"
    />
    <span class="text-body-medium">{{ filter.name }}</span>
  </div>
</template>
