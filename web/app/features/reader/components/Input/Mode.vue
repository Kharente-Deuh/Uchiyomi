<script setup lang="ts">
import type { ReadingMode } from '../../types'

defineProps<{ disabled?: boolean }>()
const mode = defineModel<ReadingMode>({ required: true })

const { t } = useI18n()

const items: { label: string, value: ReadingMode, icon: string }[] = [
  {
    label: t('reader.mode.pagedLtr'),
    value: 'paged-ltr',
    icon: 'fa6-regular:circle-right',
  },
  {
    label: t('reader.mode.pagedRtl'),
    value: 'paged-rtl',
    icon: 'fa6-regular:circle-left',
  },
  {
    label: t('reader.mode.webtoon'),
    value: 'webtoon',
    icon: 'uchi:continuous-vertical',
  },
]
</script>

<template>
  <div class="d-flex flex-column ga-2">
    <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('reader.mode.title') }}</span>
    <div class="d-flex flex-wrap ga-4">
      <VBtn
        v-for="(item, i) in items"
        :key="i"
        :value="item.value"
        :prepend-icon="item.icon"
        :text="item.label"
        :disabled
        size="large"
        class="text-uppercase"
        density="comfortable"
        :class="{ 'border-thin-primary': mode !== item.value }"
        :variant="mode === item.value ? 'flat' : 'outlined'"
        @click="mode = item.value"
      />
    </div>
  </div>
</template>
