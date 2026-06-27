<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  version: string
  hasUpdate: boolean
  size?: 'small' | 'default'
}>()
</script>

<template>
  <AtomChip
    :text="`v${version}`"
    :color="hasUpdate ? 'warning' : 'secondary'"
    :size
  />
</template>
