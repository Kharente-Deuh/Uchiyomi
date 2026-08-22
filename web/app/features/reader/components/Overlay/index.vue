<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'

defineProps<{
  comic: Comic
  chapter: DetailedChapter
}>()

const show = defineModel<boolean>({ required: true })
</script>

<template>
  <VFadeTransition>
    <ReaderOverlayHeader
      v-show="show"
      :comic="comic"
      :chapter="chapter"
    />
  </VFadeTransition>

  <VFadeTransition>
    <ReaderOverlayFooter
      v-show="show"
      :comic="comic"
      :chapter="chapter"
    />
  </VFadeTransition>
</template>
