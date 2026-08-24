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

async function retryDownload(): Promise<void> {
  const response = await comicsApi.retryChaptersDownload(props.comicId, retryableDownloadChaptersIds.value)
  if (!response.success) {
    console.error('comicsApi.retryChaptersDownload', response.error)
    toast.error(t('error.unknown'))
  }
}

async function doAction(action: 'retryDownload' | 'setRead' | 'setUnread'): Promise<void> {
  loading.value = true

  switch (action) {
    case 'retryDownload':
      await retryDownload()
      break
    case 'setRead':
      await setProgress(unreadChaptersIds.value, true)
      break
    case 'setUnread':
      await setProgress(readChaptersIds.value, false)
      break
  }

  loading.value = false

  chapters.value = []

  emit('refetchChapters', action !== 'retryDownload')
}

interface ListItem {
  disabled: boolean
  title: string
  prependIcon: string
  onClick: () => void
}

const items = computed((): ListItem[] => [
  {
    title: t('comic.chapter.download.retryMany', { count: retryableDownloadChaptersIds.value.length }),
    prependIcon: 'fa6-solid:exclamation',
    disabled: retryableDownloadChaptersIds.value.length === 0,
    onClick: () => doAction('retryDownload'),
  },
  {
    title: t('comic.chapter.setReadMany', { count: unreadChaptersIds.value.length }),
    prependIcon: 'fa6-solid:eye',
    disabled: unreadChaptersIds.value.length === 0,
    onClick: () => doAction('setRead'),
  },
  {
    title: t('comic.chapter.setUnreadMany', { count: readChaptersIds.value.length }),
    prependIcon: 'fa6-solid:eye-low-vision',
    disabled: readChaptersIds.value.length === 0,
    onClick: () => doAction('setUnread'),
  },
])
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
          v-for="({ onClick, ...item }, index) in items"
          :key="index"
          v-bind="item"
          density="compact"
          @click="onClick"
        />
      </VList>
    </VMenu>
  </VBtn>
</template>
