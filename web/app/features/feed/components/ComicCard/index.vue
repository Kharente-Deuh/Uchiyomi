<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { FeedItem } from '../../types'
import defaultCover from '~/assets/images/default/comic-cover.webp'

const props = defineProps<{ item: FeedItem }>()

const src = ref<string>(props.item.cover)
watch(() => props.item.cover, () => {
  src.value = props.item.cover
})
</script>

<template>
  <!-- <AtomLink :to="`/comic/${item.id}`"> -->
  <AtomLink>
    <div class="feed-comic-card text-truncate">
      <VImg
        cover
        :src
        rounded="lg"
        :lazy-src="defaultCover"
        aspect-ratio="2/3"
        width="100%"
        class="border-thin rounded-lg position-relative transition-smooth"
        @error="src = defaultCover"
      />
      <div class="h-100 text-truncate d-flex flex-column ga-2">
        <div class="d-flex ga-2 text-truncate justify-space-between align-center">
          <span class="text-title-large font-weight-bold feed-comic-card-title text-truncate">
            {{ item.title }}
          </span>
          <ComicsIconStatus
            v-if="(item.status as string) !== ''"
            :status="item.status"
            with-background
          />
        </div>
        <div class="d-flex flex-column ga-2 justify-space-between">
          <FeedComicCardChapter
            v-for="(chapter, index) in item.latestChapters"
            :key="index"
            :chapter
          />
        </div>
      </div>
    </div>
  </AtomLink>
</template>

<style lang="scss">
.feed-comic-card {
  display: grid;
  grid-template-columns: 25% auto;
  gap: 1.5rem;

  &:hover {
    .v-img {
      transition: all 0.2s ease-in-out;
      border: solid 1px rgb(var(--v-theme-primary));
    }

    .feed-comic-card-title {
      color: rgb(var(--v-theme-primary));
    }
  }
}
</style>
