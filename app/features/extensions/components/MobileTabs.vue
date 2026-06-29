<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
type ExtensionMobileModel = 'sources' | 'series'
const modelValue = defineModel<ExtensionMobileModel>({ required: true })
</script>

<template>
  <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; border-radius: 12px;" class="border-thin bg-surface mx-4 pa-2">
    <VBtn
      :text="$t('sources.title')"
      :color="modelValue === 'sources' ? 'primary' : 'secondary'"
      :variant="modelValue === 'sources' ? 'flat' : 'text'"
      @click="modelValue = 'sources'"
    />

    <VBtn
      :text="$t('series.title')"
      :color="modelValue === 'series' ? 'primary' : 'secondary'"
      :variant="modelValue === 'series' ? 'flat' : 'text'"
      @click="modelValue = 'series'"
    />
  </div>
</template>

<style lang="scss">
.extensions-mobile-tabs {
  &-tab {
    text-align: center;
    font-weight: bold;
    width: 100%;

    &--active {
      color: rgb(var(--v-theme-primary));
      border-bottom: 2px solid rgb(var(--v-theme-primary));
    }

    &-first {
      border-top-left-radius: 12px;
      border-bottom-left-radius: 12px;
    }

    &-last {
      border-top-right-radius: 12px;
      border-bottom-right-radius: 12px;
    }
  }
}
</style>
