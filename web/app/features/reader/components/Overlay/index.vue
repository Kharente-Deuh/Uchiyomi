<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'

defineProps<{
  comic: Comic
  chapter: DetailedChapter
  doublePage?: boolean
}>()

const show = defineModel<boolean>({ required: true })
const page = defineModel<number>('page', { required: true })
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
    <ReaderOverlayRail
      v-show="show"
      v-model:page="page"
      :chapter="chapter"
      :double-page="doublePage"
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
