<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { KingOfShojoComicInfos, KingOfShojoSearchItem } from '../../types'
import type { ComicStatus } from '~/features/comics/types'
import { KING_OF_SHOJO_SOURCE_NAME } from '~/constants'

const props = defineProps<{
  comic: KingOfShojoSearchItem | KingOfShojoComicInfos
  loading?: boolean
}>()

defineEmits<{ toggle: [] }>()

const COMIC_STATUSES: ComicStatus[] = ['ongoing', 'completed', 'hiatus', 'cancelled']

function isComicStatus(value: string): value is ComicStatus {
  return COMIC_STATUSES.includes(value as ComicStatus)
}

const api = createKingOfShojoApi()

const status = ref<ComicStatus | undefined>(isComicStatus(props.comic.status) ? props.comic.status : undefined)
const chapterCount = ref(props.comic.chapterCount)
const fetchInfosLoading = ref(false)

onMounted(async () => {
  fetchInfosLoading.value = true
  const res = await api.getInfosBySlug(props.comic.slug)

  if (res.success) {
    status.value = res.data.status
    chapterCount.value = res.data.chapterCount
  } else {
    console.error('api.getInfosBySlug', res.error)
  }

  fetchInfosLoading.value = false
})
</script>

<template>
  <SourcesCardComic
    :internal-id="comic.internalId"
    :status
    :to="`/browse/sources/${KING_OF_SHOJO_SOURCE_NAME}/${comic.slug}`"
    :cover="comic.cover"
    :title="comic.title"
    :chapter-count
    :loading
    :chapter-count-loading="fetchInfosLoading"
    :status-loading="fetchInfosLoading"
    @toggle="$emit('toggle')"
  />
</template>
