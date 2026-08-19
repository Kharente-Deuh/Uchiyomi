<script setup lang="ts">
export type PageMode = 'single' | 'double'

const props = defineProps<{
  disabledOptions?: PageMode[]
  disabled?: boolean
}>()
const scale = defineModel<PageMode>({ required: true })

const { t } = useI18n()

const items: { label: string, value: PageMode, icon: string }[] = [
  {
    label: t('reader.pageMode.singlePage'),
    value: 'single',
    icon: 'fa6-regular:file',
  },
  {
    label: t('reader.pageMode.doublePage'),
    value: 'double',
    icon: 'fa6-solid:book-open',
  },
]

watch(() => props.disabledOptions, (value) => {
  if (!value?.length) {
    return
  }

  if (value.includes(scale.value)) {
    scale.value = items.find(item => !value.includes(item.value))?.value as PageMode
  }
})

function isItemDisabled(item: PageMode): boolean {
  return props.disabledOptions?.includes(item) || props.disabled
}
</script>

<template>
  <div class="d-flex flex-column ga-2">
    <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('reader.pageMode.title') }}</span>
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
