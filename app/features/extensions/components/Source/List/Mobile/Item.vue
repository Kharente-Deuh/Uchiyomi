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
    class="d-flex ga-3 align-center prevent-select source-card-mobile justify-space-between w-100 px-3"
    :class="{
      'source-card-mobile--active': source.isEnabled,
      'bg-surface': !source.isEnabled,
    }"
    @click="isEnabled = !isEnabled"
  >
    <div class="d-flex align-center ga-4 text-truncate">
      <ExtensionsSourceChipLang :lang="source.lang" :enabled="source.isEnabled" />
      <span class="font-weight-bold text-capitalize text-truncate" :class="{ 'text-primary': source.isEnabled }">{{ displayableLang }}</span>
    </div>
    <div v-if="canManage" class="d-flex align-center ga-3">
      <VSwitch
        v-model="isEnabled"
        class="d-inline-block"
        inset
        theme="light"
        density="comfortable"
        size="x-small"
        color="primary"
        :disabled="loading"
        hide-details
      />
    </div>
  </div>
</template>

<style lang="scss">
.source-card-mobile {
  &--active {
    background-color: rgb(var(--v-theme-primary), 0.05);
  }
}
</style>
