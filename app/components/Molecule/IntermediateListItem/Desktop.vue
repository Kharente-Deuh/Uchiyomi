<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
defineProps<{
  title: string
  icon?: string
  subtitle?: string
}>()
</script>

<template>
  <div
    class="border-thin elevation-down pa-4 w-100 h-100 bg-surface"
    style="border-radius: 12px"
  >
    <div class="d-flex align-center ga-3 text-truncate mb-2">
      <VIcon :icon size="x-small" />
      <span class="text-body-large font-weight-bold text-truncate">{{ title }}</span>
    </div>
    <span class="text-label-medium text-medium-emphasis font-weight-medium" style="text-wrap: pretty;">{{ subtitle }}</span>
  </div>
</template>
