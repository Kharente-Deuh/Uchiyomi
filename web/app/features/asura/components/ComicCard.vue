<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraSearchItem } from '../types'
import defaultCover from '~/assets/images/default/comic-cover.webp'

const props = defineProps<{
  comic: AsuraSearchItem
  loading?: boolean
}>()
defineEmits<{ toggle: [] }>()

const src = ref<string>(props.comic.cover)
watch(() => props.comic.cover, (newVal) => {
  src.value = newVal
}, { immediate: true })
</script>

<template>
  <VHover>
    <template #default="{ isHovering, props: hoverProps }">
      <AtomLink :to="`/browse/sources/asura/${comic.slug}`" v-bind="hoverProps">
        <div class="d-flex flex-column comic-card w-100 h-100">
          <VImg
            cover
            :src
            rounded="lg"
            aspect-ratio="2/3"
            :lazy-src="defaultCover"
            width="100%"
            class="border-thin rounded-lg position-relative"
            :class="{ 'cover-in-library': comic.internalId }"
            @error="src = defaultCover"
          >
            <div class="d-flex w-100 justify-space-between position-absolute pa-2 top-0 lft-0" style="z-index: 1;">
              <AsuraBtnDelete
                v-if="comic.internalId"
                :mode="isHovering ? 'btn' : 'label'"
                @click.prevent="$emit('toggle')"
              />
              <AsuraBtnAdd
                v-else
                :is-in-library="!!comic.internalId"
                :loading
                class="add-library-btn"
                :class="{ 'add-library-btn-loading': loading }"
                @click.prevent="$emit('toggle')"
              />
              <ComicsIconStatus :status="comic.status" with-background />
            </div>
          </VImg>
          <span class="text-body-large font-weight-bold text-truncate mt-2 comic-card-title">{{ comic.title }}</span>
          <span class="text-body-medium text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: comic.chapterCount }) }}</span>
        </div>
      </AtomLink>
    </template>
  </VHover>
</template>

<style lang="scss">
.comic-card {
  .add-library-btn {
    opacity: 0;

    &-loading {
      opacity: 1;
    }
  }

  .in-library-label {
    padding: 0.3rem 0.4rem;
    backdrop-filter: blur(2px);
    background-color: rgba(var(--v-theme-surface), 0.7);

    &--compact {
      padding: 0.4rem;
    }
  }

  .cover-in-library {
    img {
      opacity: 0.5;
    }
  }

  &:hover {
    .in-library-label {
      opacity: 0;
    }

    .add-library-btn {
      opacity: 1;
    }

    .v-img {
      transition: all 0.2s ease-in-out;
      border: solid 1px rgb(var(--v-theme-primary));
    }

    .comic-card-title {
      transition: all 0.2s ease-in-out;
      color: rgb(var(--v-theme-primary));
    }
  }
}
</style>
