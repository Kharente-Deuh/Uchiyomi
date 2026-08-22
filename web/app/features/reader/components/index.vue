<script setup lang="ts">
import type { ReaderSettings } from '../types'
import type { DetailedChapter } from '~/features/chapters/types'
import type { Comic } from '~/features/comics/types'

const props = defineProps<{
  comic: Comic
  chapter: DetailedChapter
  settings: ReaderSettings
  retryDownloadLoading: boolean
}>()

defineEmits<{ retryDownload: [] }>()

const showOverlay = ref(false)

watch(() => props.chapter.download, (value) => {
  if (value !== 100 && !showOverlay.value) {
    showOverlay.value = true
  }
}, { immediate: true })
</script>

<template>
  <div class="h-screen w-screen position-relative">
    <ReaderOverlay
      v-model="showOverlay"
      :comic
      :chapter
    />

    <div v-if="chapter.download !== 100" class="d-flex flex-column w-100 h-100 justify-center align-center ga-4">
      <VProgressCircular
        v-if="chapter.download !== -1"
        :indeterminate="chapter.download === 0"
        :model-value="chapter.download"
        size="48"
        color="primary"
      />
      <VIcon
        v-else
        icon="fa6-solid:circle-exclamation"
        size="48"
        color="error"
      />

      <span class="text-body-large text-medium-emphasis">{{ chapter.download === -1
        ? $t('comic.chapter.download.error')
        : $t('comic.chapter.download.loading') }}</span>

      <div class="d-flex ga-4 flex-wrap">
        <VBtn
          v-if="chapter.download === -1"
          color="error"
          class="border-thin-error"
          :text="$t('comic.chapter.download.retry')"
          :loading="retryDownloadLoading"
          @click="$emit('retryDownload')"
        />
        <AtomLink :to="`/comic/${comic.id}`">
          <VBtn
            :text="$t('comic.chapter.exitToComic')"
            color="secondary"
            variant="flat"
          />
        </AtomLink>
      </div>
    </div>

    <ReaderModePaged
      v-if="settings.readingMode === 'paged-rtl' || settings.readingMode === 'paged-ltr'"
      :comic
      :chapter
      :settings
    />
  </div>
</template>
