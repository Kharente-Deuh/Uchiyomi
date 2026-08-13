<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicStatus } from '~/features/comics/types'

defineProps<{ disabled?: boolean }>()

const status = defineModel<ComicStatus | undefined>({ required: true })
const { t } = useI18n()

const items: { value: ComicStatus, title: string }[] = [
  {
    value: 'ongoing',
    title: t('sources.asura.status.ongoing'),
  },
  {
    value: 'completed',
    title: t('sources.asura.status.completed'),
  },
  {
    value: 'hiatus',
    title: t('sources.asura.status.hiatus'),
  },
  {
    value: 'cancelled',
    title: t('sources.asura.status.cancelled'),
  },
]
</script>

<template>
  <VSelect
    width="12rem"
    :model="status"
    :items="items"
    density="compact"
    class="border-thin"
    style="border-radius: 12px;"
    hide-details
    :label="$t('sources.asura.status.label')"
    clearable
    :disabled
    @update:model-value="status = $event || undefined"
  >
    <template v-if="status" #selection="{ item }">
      <div class="d-flex ga-4 align-center ga-3 text-truncate">
        <ComicsIconStatus :status="item.value" />
        <span class="text-body-medium text-truncate">{{ item.title }}</span>
      </div>
    </template>
    <template #item="{ item, props: itemProps }">
      <VListItem density="compact" v-bind="itemProps">
        <template #prepend>
          <ComicsIconStatus :status="item.value" />
        </template>
      </VListItem>
    </template>
  </VSelect>
</template>
