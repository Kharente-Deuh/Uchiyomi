<script setup lang="ts">
import type { Comic } from '../../../comics/types'
import type { DetailedChapter } from '~/features/chapters/types'
import type { ReaderSettings } from '~/features/reader/types'

defineProps<{
  comic: Comic
  chapter: DetailedChapter
  settings: ReaderSettings
}>()

function clickLeft(): void {}

function clickRight(): void {}
</script>

<template>
  <div
    class="position-absolute top-0 left-0 h-screen w-33"
    style="z-index: 1;"
    @click="clickLeft"
  />

  <div
    class="position-absolute top-0 right-0 h-screen w-33"
    style="z-index: 1;"
    @click="clickRight"
  />
</template>
