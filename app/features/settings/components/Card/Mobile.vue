<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
defineProps<{
  title: string
  adminOnly?: boolean
  loading?: boolean
}>()
</script>

<template>
  <VCard
    :loading
    :disabled="loading"
    class="d-flex flex-column bg-background elevation-0 pa-0"
  >
    <div class="d-flex align-center justify-space-between ga-4 px-6 py-1 ">
      <span class=" text-medium-emphasis text-body-small text-uppercase py-1" style="letter-spacing: 0.08em;">
        {{ title }}
      </span>
      <VChip
        v-if="adminOnly"
        :text="$t('settings.adminOnly')"
        class="rounded-lg font-title px-2 text-label-medium"
        color="secondary"
        density="compact"
      />
    </div>
    <VDivider />
    <slot />
  </VCard>
</template>
