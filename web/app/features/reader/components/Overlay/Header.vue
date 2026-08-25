<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../../../comics/types'
import type { Chapter } from '~/features/chapters/types'

defineProps<{
  comic: Comic
  chapter: Chapter
}>()

const { smAndDown } = useDisplay()
</script>

<template>
  <div
    class="d-flex align-center w-100 bg-surface border-b-thin pa-4 position-fixed top-0 left-0 right-0"
    style="z-index: 2;"
  >
    <div
      class="d-flex ga-4 justify-space-between align-center w-100 mx-auto "
      :style="smAndDown ? 'padding-top: env(safe-area-inset-top, 0px);' : ''"
      style="max-width: 80rem;"
    >
      <div class="d-flex ga-4 align-center text-truncate">
        <AtomLink :to="`/comic/${comic.id}`">
          <VBtn
            icon="fa6-solid:angle-left"
            variant="text"
            color="grey"
          />
        </AtomLink>

        <div class="d-flex ga-4 align-center text-truncate">
          <VImg
            aspect-ratio="2/3"
            :src="comic.cover"
            cover
            class="rounded-lg"
            height="100%"
            width="40"
          />
          <div class="d-flex flex-column text-truncate">
            <span class="text-body-large text-truncate text-medium-emphasis">
              {{ comic.title }}
            </span>
            <div class="d-flex ga-4 align-center">
              <span class="text-body-large font-weight-bold">
                {{ $t('feed.chapter.number', { number: chapter.number }) }}
              </span>

              <span class="text-body-large text-truncate text-medium-emphasis">
                {{ chapter.title }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <AtomLink to="/settings/reader">
        <VBtn
          v-tooltip="$t('settings.reader.title')"
          icon="fa6-solid:gear"
          variant="text"
          color="grey"
        />
      </AtomLink>
    </div>
  </div>
</template>
