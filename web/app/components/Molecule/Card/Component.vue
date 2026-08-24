<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  title: string
  icon: string
  disabled?: boolean
  loading?: boolean
}>()
</script>

<template>
  <VCard
    style="border-radius: 16px;"
    class="pa-4 border-thin"
    :disabled
    :loading
  >
    <div class="d-flex ga-4 align-center mb-6">
      <VIcon :icon color="secondary" />
      <div class="d-flex flex-column ga-1">
        <span class="font-title text-title-large">{{ title }}</span>
        <slot name="subtitle" />
      </div>
    </div>
    <slot />
  </VCard>
</template>
