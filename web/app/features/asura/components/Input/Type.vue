<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicType } from '~/features/comics/types'

defineProps<{ disabled?: boolean }>()
const sort = defineModel<ComicType | undefined>({ required: true })
const { t } = useI18n()

const items: { value: ComicType, title: string }[] = [
  {
    value: 'manhwa',
    title: t('sources.asura.type.manhwa'),
  },
  {
    value: 'mangatoon',
    title: t('sources.asura.type.mangatoon'),
  },
  {
    value: 'manga',
    title: t('sources.asura.type.manga'),
  },
  {
    value: 'manhua',
    title: t('sources.asura.type.manhua'),
  },
]
</script>

<template>
  <VSelect
    density="compact"
    width="10rem"
    hide-details
    class="border-thin"
    style="border-radius: 12px"
    :disabled
    :model="sort"
    :items="items"
    :label="$t('sources.asura.type.label')"
    clearable
    @update:model-value="sort = $event || undefined"
  />
</template>
