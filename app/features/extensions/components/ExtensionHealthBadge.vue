<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{ isHealthy?: boolean }>()
</script>

<template>
  <VChip
    v-if="isHealthy !== undefined"
    :color="isHealthy ? 'success' : 'error'"
    size="small"
    variant="tonal"
  >
    {{ isHealthy ? $t('extensions.health.ok') : $t('extensions.health.error') }}
  </VChip>
</template>
