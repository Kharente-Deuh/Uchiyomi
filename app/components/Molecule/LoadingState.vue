<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ message?: string }>()
</script>

<template>
  <div class="loading-state d-flex flex-column align-center justify-center text-center pa-8">
    <VProgressCircular
      indeterminate
      color="primary"
      size="48"
      class="mb-4"
    />
    <p class="text-medium-emphasis">
      {{ message ?? $t('states.loading') }}
    </p>
  </div>
</template>
