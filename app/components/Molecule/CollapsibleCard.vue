<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
  loading?: boolean
}>()

const isOpen = ref(false)
</script>

<template>
  <div class="w-100 d-flex flex-column ga-4 pa-4 transition-smooth">
    <div class="d-flex justify-space-between align-center text-truncate">
      <span class="text-title-large font-weight-bold text-truncate">{{ title }}</span>
      <VBtn
        :icon="isOpen ? 'fa6-solid:chevron-up' : 'fa6-solid:chevron-down'"
        color="secondary"
        size="x-small"
        variant="text"
        :loading
        @click="isOpen = !isOpen"
      />
    </div>

    <div v-show="isOpen" class="w-100 transition-smooth">
      <slot />
    </div>
  </div>
</template>
