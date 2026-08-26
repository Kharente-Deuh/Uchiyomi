<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { ComicStatus } from '~/features/comics/types'
import defaultCover from '~/assets/images/default/comic-cover.webp'

const props = defineProps<{
  internalId?: string
  status?: ComicStatus
  to: RouteLocationRaw
  cover: string
  title: string
  chapterCount: number
  chapterCountLoading?: boolean
  loading?: boolean
  statusLoading?: boolean
}>()

defineEmits<{ toggle: [] }>()

const src = ref<string>(props.cover)
watch(() => props.cover, (newVal) => {
  src.value = newVal
}, { immediate: true })

const canHover = useMediaQuery('(hover: hover) and (pointer: fine)')
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
            :class="{ 'cover-in-library': internalId }"
            @error="src = defaultCover"
          >
            <div class="d-flex w-100 justify-space-between position-absolute pa-2 top-0 left-0" style="z-index: 1;">
              <AsuraScansBtnDelete
                v-if="internalId"
                :mode="isHovering ? 'btn' : 'label'"
                @click.prevent="$emit('toggle')"
              />
              <AsuraScansBtnAdd
                v-else
                :is-in-library="!!internalId"
                :loading
                class="add-library-btn"
                :class="{ 'add-library-btn-loading': loading }"
                @click.prevent="$emit('toggle')"
              />
              <ComicsIconStatus
                :status
                with-background
                :loading="statusLoading"
              />
            </div>
          </VImg>
          <span class="text-body-large font-weight-bold text-truncate mt-2 comic-card-title">{{ title }}</span>
          <VSkeletonLoader
            v-if="chapterCountLoading"
            type="text"
            class="mt-1 w-33"
          />
          <span v-else class="text-body-medium text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: chapterCount }) }}</span>
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
