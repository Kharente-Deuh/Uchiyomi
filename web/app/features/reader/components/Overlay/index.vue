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
  <VExpandTransition>
    <ReaderOverlayHeader
      v-show="show"
      :comic="comic"
      :chapter="chapter"
    />
  </VExpandTransition>

  <VExpandTransition>
    <ReaderOverlayFooter
      v-show="show"
      :comic="comic"
      :chapter="chapter"
    />
  </VExpandTransition>
</template>
