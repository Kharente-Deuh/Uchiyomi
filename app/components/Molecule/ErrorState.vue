<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ message?: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div class="error-state d-flex flex-column align-center justify-center text-center pa-8">
    <VIcon
      icon="fa6-solid:triangle-exclamation"
      size="48"
      color="error"
      class="mb-4"
    />
    <p class="text-medium-emphasis mb-4">
      {{ message ?? $t('states.error') }}
    </p>
    <VBtn
      color="primary"
      variant="tonal"
      @click="$emit('retry')"
    >
      {{ $t('states.retry') }}
    </VBtn>
  </div>
</template>
