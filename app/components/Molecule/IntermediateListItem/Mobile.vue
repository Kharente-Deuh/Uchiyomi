<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  title: string
  icon?: string
  subtitle?: string
  last?: boolean
}>()
</script>

<template>
  <div
    class="border-t-thin elevation-down px-4 py-2 w-100 bg-surface d-flex justify-space-between align-center"
    :class="{ 'border-b-thin': last }"
  >
    <div>
      <div class="d-flex align-center ga-3 text-truncate mb-1">
        <VIcon :icon size="x-small" />
        <span class="text-body-large font-weight-bold text-truncate">{{ title }}</span>
      </div>
      <span class="text-label-medium text-medium-emphasis font-weight-medium" style="text-wrap: pretty;">{{ subtitle }}</span>
    </div>
    <VIcon icon="fa6-solid:chevron-right" size="x-small" />
  </div>
</template>
