<script setup lang="ts">
import type { PageScale } from '../../types'

const props = defineProps<{
  disabledOptions?: PageScale[]
  disabled?: boolean
}>()
const scale = defineModel<PageScale>({ required: true })

const { t } = useI18n()

const items: { label: string, value: PageScale, icon: string }[] = [
  {
    label: t('reader.pageScale.fitWidth'),
    value: 'fit-width',
    icon: 'fa6-solid:arrows-left-right-to-line',
  },
  {
    label: t('reader.pageScale.fitHeight'),
    value: 'fit-height',
    icon: 'uchi:screen-height',
  },
  {
    label: t('reader.pageScale.fitScreen'),
    value: 'fit-screen',
    icon: 'uchi:screen-fit',
  },
]

watch(() => props.disabledOptions, (value) => {
  if (!value?.length) {
    return
  }

  if (value.includes(scale.value)) {
    scale.value = items.find(item => !value.includes(item.value))?.value as PageScale
  }
})

function isItemDisabled(item: PageScale): boolean {
  return props.disabledOptions?.includes(item) || props.disabled
}
</script>

<template>
  <div class="d-flex flex-column ga-2">
    <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('reader.pageScale.title') }}</span>
    <div class="d-flex flex-wrap ga-4">
      <VBtn
        v-for="(item, i) in items"
        :key="i"
        :value="item.value"
        :prepend-icon="item.icon"
        :text="item.label"
        :color="isItemDisabled(item.value) ? 'grey' : 'primary'"
        size="large"
        :disabled="isItemDisabled(item.value)"
        class="text-uppercase"
        density="comfortable"
        :variant="scale === item.value ? 'flat' : 'outlined'"
        @click="scale = item.value"
      />
    </div>
  </div>
</template>
