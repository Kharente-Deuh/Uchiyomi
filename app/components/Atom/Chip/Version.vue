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
