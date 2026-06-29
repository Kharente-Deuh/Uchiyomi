<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<{
  size?: 'small' | 'default'
}>()
</script>

<template>
  <AtomChip
    text="18+"
    color="error"
    :size
  />
</template>
