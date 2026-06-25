<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()

const { mobile } = useDisplay()
</script>

<template>
  <div v-if="!mobile" class="d-flex align-center justify-space-between ga-4 align-center">
    <div class="d-flex flex-column text-truncate">
      <span class="text-body-medium font-weight-bold text-truncate">{{ title }}</span>
      <span v-if="subtitle" class="text-label-small text-medium-emphasis text-truncate">{{ subtitle }}</span>
    </div>

    <slot />
  </div>
  <div v-else class="d-flex align-center px-4  py-3 ga-4 justify-space-between bg-surface">
    <span class="text-body-medium font-weight-bold text-truncate">{{ title }}</span>
    <slot />
  </div>
</template>
