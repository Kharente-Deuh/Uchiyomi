<script setup lang="ts">
import type { SearchComicSort } from '~/features/comics/types'

defineProps<{ disabled?: boolean }>()
const sort = defineModel<SearchComicSort>({ required: true })
const { t } = useI18n()

const items: { value: SearchComicSort, title: string }[] = [
  {
    value: 'title',
    title: t('library.sort.title'),
  },
  {
    value: 'addedAt',
    title: t('library.sort.addedAt'),
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
