<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { SourceSearchItem } from '../../types'
import type { ComicSource } from '~/features/comics/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'

const props = defineProps<{
  sourceId: ComicSource
  comic: SourceSearchItem
  loading?: boolean
  chapterCountLoading?: boolean
  statusLoading?: boolean
}>()

defineEmits<{ toggle: [] }>()

const src = ref<string>(props.comic.cover)
watch(() => props.comic.cover, (newVal) => {
  src.value = newVal
}, { immediate: true })

const canHover = useMediaQuery('(hover: hover) and (pointer: fine)')

const to = computed<RouteLocationRaw>(() => `/browse/sources/${props.sourceId}/${props.comic.slug}`)
</script>

<template>
  <VHover :disabled="!canHover">
    <template #default="{ isHovering, props: hoverProps }">
      <AtomLink :to v-bind="hoverProps">
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
            <div class="d-flex w-100 justify-space-between position-absolute pa-2 top-0 left-0" style="z-index: 1;">
              <SourcesBtnDelete
                v-if="comic.internalId"
                :source-id="sourceId"
                :mode="isHovering ? 'btn' : 'label'"
                @click.prevent="$emit('toggle')"
              />
              <SourcesBtnAdd
                v-else
                :loading
                class="add-library-btn"
                :class="{ 'add-library-btn-loading': loading }"
                @click.prevent="$emit('toggle')"
              />
              <ComicsIconStatus
                :status="comic.status"
                with-background
                :loading="statusLoading"
              />
            </div>
          </VImg>
          <span class="text-body-large font-weight-bold text-truncate mt-2 comic-card-title">{{ comic.title }}</span>
          <VSkeletonLoader
            v-if="chapterCountLoading"
            type="text"
            class="mt-1 w-33"
          />
          <span v-else class="text-body-medium text-medium-emphasis">{{ $t('sources.comic.chaptersCount', { count: comic.chapterCount }) }}</span>
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

  @media (hover: none), (pointer: coarse) {
    .add-library-btn {
      opacity: 1;
    }
  }

  @media (hover: hover) and (pointer: fine) {
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
}
</style>
