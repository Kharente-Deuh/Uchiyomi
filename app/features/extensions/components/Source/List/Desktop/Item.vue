<script setup lang="ts">
import type { SourceDto } from '#shared/dto/extensions'
import { endomyLanguage } from '~/utils/languages.util'

const props = defineProps<{
  source: SourceDto
  loading?: boolean
  canManage?: boolean
  mobile?: boolean
}>()

const emits = defineEmits<{
  toggle: []
  settings: []
}>()

const isEnabled = computed({
  get: () => props.source.isEnabled,
  set: () => emits('toggle'),
})

const { t } = useI18n()
const displayableLang = ref<string>(props.source.lang === 'all'
  ? t('sources.all')
  : endomyLanguage(props.source.lang))
</script>

<template>
  <div
    class="d-flex ga-3 align-center w-fit prevent-select source-card elevation-down"
    :class="{
      'px-3': canManage,
      'pa-2': !canManage,
      'border-thin': !source.isEnabled,
      'source-card--active': source.isEnabled,
      'bg-surface': !source.isEnabled,
    }"
    @click="isEnabled = !isEnabled"
  >
    <ExtensionsSourceChipLang :lang="source.lang" :enabled="source.isEnabled" />
    <span class="font-weight-bold text-capitalize text-nowrap" :class="{ 'text-primary': source.isEnabled }">{{ displayableLang }}</span>
    <VSwitch
      v-if="canManage"
      v-model="isEnabled"
      inset
      class="d-inline-block"
      theme="light"
      density="comfortable"
      size="x-small"
      color="primary"
      :disabled="loading"
      hide-details
    />
  </div>
</template>

<style lang="scss">
.source-card {
  border-radius: 10px;

  &--active {
    background-color: rgb(var(--v-theme-primary), 0.05);
    border: 1px solid rgb(var(--v-theme-primary));
  }
}
</style>
