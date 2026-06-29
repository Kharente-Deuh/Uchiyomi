<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceFilterDto } from '#shared/dto/catalogue/source-filters.dto'
import type { FilterDraftValue } from '~/features/extensions/composables/source-filters.composable'
import { filterPath } from '~/features/extensions/composables/source-filters.composable'

type LeafFilter = Exclude<SourceFilterDto, { type: 'group' }>

const props = defineProps<{
  definitions: SourceFilterDto[]
}>()
const emit = defineEmits<{ apply: [], reset: [] }>()
const show = defineModel<boolean>({ required: true })
const draft = defineModel<Record<string, FilterDraftValue>>('draft', { required: true })

/** Cast a filter to a leaf type for use in Control — groups are handled inline above. */
function asLeaf(filter: SourceFilterDto): LeafFilter {
  return filter as LeafFilter
}

function onApply(): void {
  emit('apply')
  show.value = false
}
</script>

<template>
  <OrganismModal
    v-model="show"
    prepend-icon="fa6-solid:filter"
    :title="$t('sources.series.filters.title')"
    :submit-text="$t('sources.series.filters.apply')"
    @submit="onApply"
    @cancel="show = false"
  >
    <template
      v-for="filter in props.definitions"
      :key="filter.position"
    >
      <template v-if="filter.type === 'group'">
        <div class="text-body-small text-medium-emphasis text-uppercase mt-2">
          {{ filter.name }}
        </div>
        <ExtensionsSeriesFiltersControl
          v-for="child in filter.filters"
          :key="`${filter.position}.${child.position}`"
          v-model="draft[filterPath(child.position, filter.position)]"
          :filter="asLeaf(child)"
        />
      </template>
      <ExtensionsSeriesFiltersControl
        v-else
        v-model="draft[filterPath(filter.position)]"
        :filter="asLeaf(filter)"
      />
    </template>

    <template #actions>
      <VBtn
        variant="text"
        block
        :text="$t('sources.series.filters.reset')"
        @click="emit('reset')"
      />
    </template>
  </OrganismModal>
</template>
