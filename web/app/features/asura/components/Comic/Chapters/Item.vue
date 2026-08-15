<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicChapter } from '~/features/asura/types'
import { formatRelativeTime } from '~/utils/date.utils'

const props = defineProps<{
  chapter: AsuraComicChapter
  comicOriginUrl: string
}>()

const { locale } = useI18n()

const date = computed(() => {
  const aMonthAgo = new Date()
  aMonthAgo.setMonth(aMonthAgo.getMonth() - 1)

  if (props.chapter.publishedAt < aMonthAgo) {
    return props.chapter.publishedAt.toLocaleDateString(locale.value, { dateStyle: 'medium' })
  }

  return formatRelativeTime(props.chapter.publishedAt, { locale: locale.value })
})

const isEarlyAccess = computed(() => props.chapter.earlyAccessUntil > new Date())

const to = computed(() => {
  if (isEarlyAccess.value) {
    return
  }

  return `${props.comicOriginUrl}/chapter/${props.chapter.number}`
})
</script>

<template>
  <AtomLink
    :to="to"
    external
    new-tab
    no-external-icon
  >
    <div
      class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center text-truncate transition-smooth"
      :class="{
        'early-access-chapter': isEarlyAccess,
        'readable-chapter': !isEarlyAccess,
      }"
    >
      <div class="d-flex align-center ga-4">
        <div v-if="isEarlyAccess" class="d-flex flex-column items-center justify-center border-thin-gold pa-2 text-body-small rounded-lg early-access-icon">
          <VIcon
            icon="fa6-solid:lock"
            size="x-small"
            color="gold"
          />
        </div>
        <div
          class="d-flex flex-wrap text-truncate"
          :class="{
            'flex-column': isEarlyAccess,
            'align-center': !isEarlyAccess,
            'ga-4': !isEarlyAccess,
          }"
        >
          <span class="text-body-large font-weight-bold">{{ $t('sources.asurascans.comic.chapter', { number: chapter.number }) }}</span>
          <span v-if="chapter.title && !isEarlyAccess" class="text-body-medium text-medium-emphasis text-truncate">{{ chapter.title }}</span>
          <span
            v-else-if="isEarlyAccess"
            class="text-body-medium text-medium-emphasis text-truncate text-gold"
          >
            {{ $t('sources.asurascans.comic.chapterUnlocksIn', { time: formatRelativeTime(chapter.earlyAccessUntil, { locale, direction: 'future' }) }) }}
          </span>
        </div>
      </div>

      <div class="d-flex align-center ga-3">
        <VProgressCircular
          v-if="chapter.internalId && chapter.download !== undefined && chapter.download >= 0 && chapter.download < 100"
          :model-value="chapter.download"
          size="18"
          :indeterminate="chapter.download === 0"
          width="2"
          color="primary"
        />
        <VIcon
          v-else-if="chapter.internalId && chapter.download === 100"
          icon="fa6-solid:check"
          size="x-small"
          color="success"
        />
        <VIcon
          v-else-if="chapter.internalId && chapter.download === -1"
          icon="fa6-solid:exclamation"
          size="x-small"
          color="error"
        />
        <span class="text-body-medium text-medium-emphasis" :class="{ 'text-gold': isEarlyAccess }">{{ date }}</span>
      </div>
    </div>
  </AtomLink>
</template>

<style lang="scss">
.readable-chapter {
  &:hover {
    color: rgb(var(--v-theme-primary));
  }
}

.early-access-chapter {
  background: linear-gradient(90deg, rgba(var(--v-theme-gold), 0.05) 0%, rgba(0, 0, 0, 0) 80%);
}

.early-access-icon {
  background-color: rgb(var(--v-theme-gold), 0.1);
}
</style>
