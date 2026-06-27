<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
  showSubtitle?: boolean
  wrap?: boolean
}>()

const { mobile } = useDisplay()
</script>

<template>
  <div
    class="d-flex align-center justify-space-between ga-4 align-center"
    :class="{
      'flex-wrap': wrap,
      'py-3 px-4 bg-surface': mobile,
    }"
  >
    <div class="d-flex flex-column">
      <span class="text-body-medium font-weight-bold text-wrap">{{ title }}</span>
      <span v-if="subtitle" class="text-label-small text-medium-emphasis text-wrap">{{ subtitle }}</span>
    </div>

    <slot />
  </div>
</template>
