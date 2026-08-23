<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { DetailedChapter } from '~/features/chapters/types'

const props = defineProps<{
  chapter: DetailedChapter
  doublePage?: boolean
}>()

const page = defineModel<number>('page', { default: 0 })

const lastIndex = computed(() => Math.max(props.chapter.pageUrls.length - 1, 0))
const step = computed(() => props.doublePage ? 2 : 1)
const currentLabel = computed(() => {
  const current = page.value + 1
  const total = props.chapter.pageUrls.length
  if (!props.doublePage || current >= total) {
    return String(current)
  }

  return `${current}–${current + 1}`
})
</script>

<template>
  <div
    class="position-fixed d-flex py-8"
    style="top: 6rem; bottom: 4.5rem; right: 1.5rem; z-index: 2;"
  >
    <div class="d-flex flex-column rounded-pill bg-surface border-thin pa-4 ga-2 h-100">
      <div data-test="rail-current" class="d-flex w-100 justify-center">
        {{ currentLabel }}
      </div>

      <VSlider
        v-model="page"
        class="rail-slider flex-grow-1"
        :min="0"
        :max="lastIndex"
        :step="step"
        direction="vertical"
        reverse
        readonly
        color="primary"
        track-color="secondary"
        show-ticks="always"
        :tick-size="4"
        :track-size="20"
        :thumb-size="4"
        hide-details
      />

      <div class="d-flex w-100 justify-center">
        {{ chapter.pageUrls.length }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.rail-slider {
  min-height: 0;
  overflow: visible;
}

.rail-slider :deep(.v-slider.v-input--vertical) {
  margin-block: 0;
}

.rail-slider :deep(.v-slider.v-input--vertical > .v-input__control) {
  min-height: 0;
  height: 100%;
  overflow: visible;
}

.rail-slider :deep(.v-slider-track) {
  border-radius: 9999px;
}

.rail-slider :deep(.v-slider-track__fill) {
  border-radius: 9999px 9999px 0 0;
}

.rail-slider :deep(.v-slider-thumb__surface) {
  width: 28px;
  height: 4px;
  border-radius: 2px;
}

.rail-slider :deep(.v-slider-thumb__surface::before),
.rail-slider :deep(.v-slider-thumb__ripple) {
  display: none;
}

.rail-slider :deep(.v-slider.v-input--vertical .v-slider-track__tick) {
  opacity: 1;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  margin-inline-start: calc((var(--v-slider-track-size) + 2px) / 2);
  transform: translate(-50%, 50%);
}
</style>
