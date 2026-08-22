<script setup lang="ts">
import type { ComicSource } from '~/features/comics/types'

const props = withDefaults(defineProps<{
  source: ComicSource
  size?: 'default' | 'small'
}>(), { size: 'default' })

const { getSourceDetails } = useSources()

const details = computed(() => getSourceDetails(props.source))
</script>

<template>
  <VImg
    v-if="size === 'small'"
    :src="details.image"
    aspect-ratio="1"
    width="31"
    class="rounded-lg"
    :style="{ border: `1px solid ${details.color}C8` }"
  />
  <div
    v-else
    class="d-flex align-center ga-3 w-fit px-2 py-1 bg-background"
    :style="{ borderRadius: '12px', border: `1px solid ${details.color}C8` }"
  >
    <VImg
      :src="details.image"
      aspect-ratio="1"
      width="20"
      class="rounded-lg"
    />
    <span class="text-body-large text-truncate">{{ details.name }}</span>
  </div>
</template>
