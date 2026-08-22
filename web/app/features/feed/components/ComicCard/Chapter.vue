<script setup lang="ts">
import type { FeedChapter } from '../../types'
import { formatRelativeTime } from '~/utils/date.utils'

const props = defineProps<{
  comicId: string
  chapter: FeedChapter
}>()

const { locale } = useI18n()

const date = computed(() => {
  return formatRelativeTime(props.chapter.earlyAccessUntil ?? props.chapter.publishedAt, { locale: locale.value })
})
</script>

<template>
  <AtomLink :to="chapter.download === 100 ? `/comic/${props.comicId}/${props.chapter.id}` : undefined">
    <div class="d-flex justify-space-between ga-2 transition-smooth align-center" :class="{ 'feed-chapter': chapter.download === 100 }">
      <div class="d-flex ga-2 align-center">
        <span
          class="text-body-large feed-chapter-title"
          style="font-size: 1.05rem;"
          :class="{ 'text-medium-emphasis': chapter.download !== 100 }"
        > {{ $t('feed.chapter.number', { number: chapter.number }) }} </span>
        <template v-if="chapter.download !== 100">
          <VIcon
            v-if="chapter.download === -1"
            icon="fa6-solid:exclamation"
            size="x-small"
            color="error"
          />
          <VProgressCircular
            v-else
            :value="chapter.download"
            size="18"
            :indeterminate="chapter.download === 0"
            width="2"
            color="primary"
          />
        </template>
      </div>
      <span class="text-body-medium text-medium-emphasis"> {{ date }} </span>
    </div>
  </AtomLink>
</template>

<style lang="scss">
.feed-chapter {
  &:hover {
    .feed-chapter-title {
      color: rgb(var(--v-theme-primary));
    }
  }
}
</style>
