<script setup lang="ts">
import type { ComicProgressContinue } from '../../types'
import type { Chapter } from '~/features/chapters/types'

const props = defineProps<{
  comicId: string
  continue?: ComicProgressContinue
  chapters: Chapter[]
  sort: 'asc' | 'desc'
}>()

const { smAndDown } = useDisplay()

const nextChapter = computed(() => {
  const chaptersCpy = props.sort === 'asc' ? props.chapters : props.chapters.toSorted((a, b) => a.number - b.number)
  const continueCpy = props.continue

  if (!continueCpy) {
    return chaptersCpy[0] as Chapter
  }

  const i = chaptersCpy.findIndex(chapter => chapter.id === continueCpy.chapterId)
  if (i === -1) {
    return
  }

  if (continueCpy.page === chaptersCpy[i]!.pagesNb) {
    if (i === chaptersCpy.length - 1) {
      return
    }

    return chaptersCpy[i + 1] as Chapter
  }

  return chaptersCpy[i] as Chapter
})

const nextChapterText = computed(() => {
  if (!nextChapter.value) {
    return $t('common.upToDate')
  }

  return nextChapter.value.number === 1 ? $t('common.start') : $t('common.continue')
})

const to = computed(() => {
  if (!nextChapter.value) {
    return undefined
  }

  return `/comic/${props.comicId}/${nextChapter.value.id}`
})

const icon = computed(() => {
  if (!nextChapter.value) {
    return 'fa6-solid:check'
  }

  return 'fa6-solid:play'
})
</script>

<template>
  <AtomLink :to>
    <VBtn
      :variant="nextChapter ? 'tonal' : 'outlined'"
      :class="{ 'border-thin-primary': nextChapter }"
      :prepend-icon="icon"
      :text="nextChapterText"
      :readonly="!nextChapter"
      :color="nextChapter ? 'primary' : 'secondary'"
      :size="smAndDown ? 'small' : undefined"
    />
  </AtomLink>
</template>
