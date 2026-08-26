<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { KingOfShojoSort } from '../../types'

defineProps<{ disabled?: boolean }>()
const sort = defineModel<KingOfShojoSort>({ required: true })
const { t } = useI18n()

const items: { value: KingOfShojoSort, title: string }[] = [
  {
    value: 'popular',
    title: t('sources.asurascans.sort.popular'),
  },
  {
    value: 'latest',
    title: t('sources.asurascans.sort.latest'),
  },
  {
    value: 'title',
    title: t('sources.asurascans.sort.title'),
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
    :label="$t('sources.asurascans.sort.label')"
  />
</template>
