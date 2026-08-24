<script setup lang="ts">
import type { Chapter } from '~/features/chapters/types'

const props = defineProps<{ comicId: string }>()
const emit = defineEmits<{ refetchChapters: [updateConitnue: boolean] }>()
const chapters = defineModel<Chapter[]>({ required: true })

const comicsApi = createComicsApi()
const { t } = useI18n()
const toast = useToast()

const readChaptersIds = computed(() => {
  const ids: string[] = []
  for (const chapter of chapters.value) {
    if (chapter.pagesNb === chapter.progress?.page) {
      ids.push(chapter.id)
    }
  }

  return ids
})

const unreadChaptersIds = computed(() => {
  const ids: string[] = []
  for (const chapter of chapters.value) {
    if (chapter.pagesNb !== chapter.progress?.page) {
      ids.push(chapter.id)
    }
  }

  return ids
})

const retryableDownloadChaptersIds = computed(() => {
  const ids: string[] = []
  for (const chapter of chapters.value) {
    if (chapter.download === -1) {
      ids.push(chapter.id)
    }
  }

  return ids
})

const showMenu = ref(false)
const loading = ref(false)

async function setProgress(chaptersIds: string[], read: boolean): Promise<void> {
  const response = await comicsApi.setChaptersProgress({
    comicId: props.comicId,
    chapterIds: chaptersIds,
    read,
  })

  if (!response.success) {
    console.error('comicsApi.setChaptersProgress', response.error)
    toast.error(t('error.unknown'))
  }
}

async function doAction(action: 'retryDownload' | 'setRead' | 'setUnread'): Promise<void> {
  loading.value = true

  if (action !== 'retryDownload') {
    await setProgress(
      action === 'setRead' ? unreadChaptersIds.value : readChaptersIds.value,
      action === 'setRead',
    )
  }

  loading.value = false

  chapters.value = []

  emit('refetchChapters', action !== 'retryDownload')
}
</script>

<template>
  <VBtn
    prepend-icon="fa6-solid:gear"
    color="surfaceVariant"
    :loading
  >
    {{ $t('common.actions') }}
    <VMenu
      v-model="showMenu"
      activator="parent"
      location="bottom center"
      offset="8"
      close-on-content-click
    >
      <VList
        density="compact"
        style="border-radius: 12px;"
        class="border-thin"
      >
        <VListItem
          :title="$t('comic.chapter.download.retryMany', { count: retryableDownloadChaptersIds.length })"
          :disabled="retryableDownloadChaptersIds.length === 0"
          prepend-icon="fa6-solid:exclamation"
          density="compact"
          @click="doAction('retryDownload')"
        />
        <VListItem
          :title="$t('comic.chapter.setReadMany', { count: unreadChaptersIds.length })"
          :disabled="unreadChaptersIds.length === 0"
          prepend-icon="fa6-solid:check"
          density="compact"
          @click="doAction('setRead')"
        />

        <VListItem
          :title="$t('comic.chapter.setUnreadMany', { count: readChaptersIds.length })"
          :disabled="readChaptersIds.length === 0"
          prepend-icon="fa6-solid:eye-low-vision"
          density="compact"
          @click="doAction('setUnread')"
        />
      </VList>
    </VMenu>
  </VBtn>
</template>
