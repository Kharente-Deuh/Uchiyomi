<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { KingOfShojoComicChapter } from '~/features/kingofshojo/types'
import { formatRelativeTime } from '~/utils/date.utils'

const props = defineProps<{
  chapter: KingOfShojoComicChapter
  comicOriginUrl: string
}>()

const { retryDownload, retryDownloadLoading } = useKingOfShojoChapters()

const { locale } = useI18n()

const date = computed(() => {
  const aMonthAgo = new Date()
  aMonthAgo.setMonth(aMonthAgo.getMonth() - 1)

  if (props.chapter.publishedAt < aMonthAgo) {
    return props.chapter.publishedAt.toLocaleDateString(locale.value, { dateStyle: 'medium' })
  }

  return formatRelativeTime(props.chapter.publishedAt, { locale: locale.value })
})

const isChapterDownloading = computed(() => props.chapter.internalId
  && props.chapter.download !== undefined
  && props.chapter.download >= 0
  && props.chapter.download < 100
  && (!props.chapter.earlyAccessUntil || props.chapter.earlyAccessUntil < new Date()))
const isChapterDownloaded = computed(() => props.chapter.internalId && props.chapter.download === 100)
const isChapterDownloadingError = computed(() => props.chapter.internalId && props.chapter.download === -1 && !retryDownloadLoading.value)

const to = computed(() => `${props.comicOriginUrl}-chapter-${String(props.chapter.number).replaceAll('.', '-')}`)
</script>

<template>
  <AtomLink
    :to
    external
    new-tab
    no-external-icon
  >
    <div class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center text-truncate transition-smooth readable-chapter">
      <div class="d-flex align-center ga-4">
        <div class="d-flex flex-wrap text-truncate align-center ga-4">
          <span class="text-body-large font-weight-bold">{{ $t('sources.kingofshojo.comic.chapter', { number: chapter.number }) }}</span>
          <span v-if="chapter.title" class="text-body-medium text-medium-emphasis text-truncate">{{ chapter.title }}</span>
        </div>
      </div>

      <div class="d-flex align-center ga-3">
        <VProgressCircular
          v-if="isChapterDownloading"
          :model-value="chapter.download"
          size="18"
          :indeterminate="chapter.download === 0"
          width="2"
          color="primary"
        />
        <VIcon
          v-if="isChapterDownloaded"
          icon="fa6-solid:check"
          size="x-small"
          color="success"
        />
        <VIcon
          v-if="isChapterDownloadingError && !retryDownloadLoading"
          v-tooltip:bottom="$t('sources.kingofshojo.comic.retryDownloadChapter.tooltip')"
          icon="fa6-solid:exclamation"
          classs="cursor-pointer"
          size="x-small"
          color="error"
          @click.prevent="retryDownload(chapter.internalId as string)"
        />
        <VProgressCircular
          v-if="isChapterDownloadingError && retryDownloadLoading"
          indeterminate
          size="18"
          width="2"
          color="error"
        />
        <span class="text-body-medium text-medium-emphasis">{{ date }}</span>
      </div>
    </div>
  </AtomLink>
</template>

<style lang="scss">
.readable-chapter {
  &:hover {
    color: rgb(var(--v-theme-primary));
    background-color: rgba(var(--v-theme-surface-variant));
  }
}
</style>
