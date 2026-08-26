<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceSort } from '../../types'
import type { ComicSource } from '~/features/comics/types'
import { getSourceConfig } from '../../config/sources.config'

const props = defineProps<{
  source: ComicSource
  disabled?: boolean
}>()

const sort = defineModel<SourceSort>({ required: true })
const { t } = useI18n()

const items = computed(() => {
  const config = getSourceConfig(props.source)
  if (!config) {
    return []
  }

  return config.allowedSorts.map(value => ({
    value,
    title: t(`sources.sort.${value}`),
  }))
})
</script>

<template>
  <VSelect
    v-model="sort"
    density="compact"
    class="border-thin"
    style="border-radius: 12px"
    width="12rem"
    :disabled
    hide-details
    :items="items"
    :label="$t('sources.sort.label')"
  />
</template>
