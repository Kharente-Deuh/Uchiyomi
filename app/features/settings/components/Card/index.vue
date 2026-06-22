<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
  adminOnly?: boolean
}>()
</script>

<template>
  <div
    class="pa-6 bg-surface d-flex flex-column ga-8 border-thin elevation-down"
    style="border-radius: 12px;"
  >
    <div class="d-flex flex-column ga-2">
      <div class="d-flex text-truncate justify-space-between ga-2 align-center">
        <span class="font-title font-weight-bold text-title-large text-truncate">{{ title }}</span>
        <VChip
          v-if="adminOnly"
          :text="$t('settings.adminOnly')"
          class="rounded-lg font-title"
          color="secondary"
        />
      </div>
      <span class="text-label-medium text-medium-emphasis text-truncate font-weight-medium text-truncate">{{ subtitle }}</span>
    </div>

    <div class="d-flex flex-column ga-4">
      <slot />
    </div>
  </div>
</template>
