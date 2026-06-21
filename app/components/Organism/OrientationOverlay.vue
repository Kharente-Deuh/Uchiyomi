<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
defineProps<{ modelValue: boolean }>()
</script>

<template>
  <VOverlay
    :model-value="modelValue"
    persistent
    scrim="background"
    class="orientation-overlay align-center justify-center"
  >
    <div class="text-center pa-8">
      <VIcon
        icon="fa6-solid:mobile-screen-button"
        size="64"
        color="primary"
        class="mb-4"
      />
      <h2 class="text-h6 mb-2">
        {{ $t('orientation.title') }}
      </h2>
      <p class="text-medium-emphasis">
        {{ $t('orientation.message') }}
      </p>
    </div>
  </VOverlay>
</template>
