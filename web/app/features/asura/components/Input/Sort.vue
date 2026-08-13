<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraSort } from '../../types'

defineProps<{ disabled?: boolean }>()
const sort = defineModel<AsuraSort>({ required: true })
const { t } = useI18n()

const items: { value: AsuraSort, title: string }[] = [
  {
    value: 'popular',
    title: t('sources.asura.sort.popular'),
  },
  {
    value: 'latest',
    title: t('sources.asura.sort.latest'),
  },
  {
    value: 'title',
    title: t('sources.asura.sort.title'),
  },
  {
    value: 'rating',
    title: t('sources.asura.sort.rating'),
  },
]
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
    :label="$t('sources.asura.sort.label')"
  />
</template>
