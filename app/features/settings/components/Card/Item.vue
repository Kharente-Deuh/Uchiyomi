<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="d-flex align-center justify-space-between ga-4 align-center">
    <div class="d-flex flex-column text-truncate">
      <span class="text-body-medium font-weight-bold text-truncate">{{ title }}</span>
      <span v-if="subtitle" class="text-label-small text-medium-emphasis text-truncate">{{ subtitle }}</span>
    </div>

    <slot />
  </div>
</template>
