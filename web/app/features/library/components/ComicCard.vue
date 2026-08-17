<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { LightComic } from '~/features/comics/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'

const props = defineProps<{ comic: LightComic }>()

const src = ref<string>(props.comic.cover)
watch(() => props.comic.cover, (newVal) => {
  src.value = newVal
}, { immediate: true })
</script>

<template>
  <AtomLink :to="`/comic/${comic.id}`">
    <div class="library-comic-card w-100 h-100 d-flex flex-column">
      <VImg
        cover
        :src
        rounded="lg"
        :lazy-src="defaultCover"
        aspect-ratio="2/3"
        width="100%"
        class="border-thin rounded-lg"
        @error="src = defaultCover"
      />
      <span class="text-body-large font-weight-bold text-truncate mt-2 library-comic-card-title">{{ comic.title }}</span>
      <span class="text-body-medium text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: comic.chapterCount }) }}</span>
    </div>
  </AtomLink>
</template>

<style lang="scss">
.library-comic-card {
  &:hover {
    .v-img {
      transition: all 0.2s ease-in-out;
      border: solid 1px rgb(var(--v-theme-primary));
    }

    .library-comic-card-title {
      transition: all 0.2s ease-in-out;
      color: rgb(var(--v-theme-primary));
    }
  }
}
</style>
