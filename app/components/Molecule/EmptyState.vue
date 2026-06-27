<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ icon?: string, title?: string, description?: string }>()
</script>

<template>
  <div class="empty-state d-flex flex-column align-center justify-center text-center pa-8">
    <VIcon
      :icon="icon ?? 'fa6-solid:inbox'"
      size="48"
      color="medium-emphasis"
      class="mb-4"
    />
    <h2 class="text-h6 mb-2">
      {{ title ?? $t('states.empty') }}
    </h2>
    <p v-if="description" class="text-medium-emphasis mb-4">
      {{ description }}
    </p>
    <slot name="action" />
  </div>
</template>
