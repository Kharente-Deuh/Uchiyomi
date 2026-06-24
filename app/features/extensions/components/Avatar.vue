<script setup lang="ts">
defineProps<{ url: string }>()
</script>

<template>
  <div style="width: 48px; height: 48px;">
    <VImg
      :src="url"
      class="rounded-lg"
      fill
      aspect-ratio="1/1"
    />
  </div>
</template>
