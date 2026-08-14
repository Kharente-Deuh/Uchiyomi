<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicChapter } from '../../../types'

const props = defineProps<{ slug: string }>()

const api = createAsuraApi()
const toast = useToast()
const { t } = useI18n()

const fetchChaptersLoading = ref(false)
const chapters = ref<AsuraComicChapter[]>([])

onMounted(() => {
  fetchChapters()
})

async function fetchChapters(): Promise<void> {
  fetchChaptersLoading.value = true

  const res = await api.getSeriesChapters(props.slug as string)

  if (res.success) {
    chapters.value = res.data
  } else {
    console.error('api.getSeriesChapters', res.error)
    toast.error(res.error.status === 404
      ? t('sources.asura.comic.chapters.error.fetch')
      : t('error.unknown'))
  }

  fetchChaptersLoading.value = false
}

const sort = ref<'asc' | 'desc'>('desc')

const filteredChapters = computed(() => chapters.value.toSorted((a, b) => {
  if (sort.value === 'asc') {
    return a.number - b.number
  }

  return b.number - a.number
}))
</script>

<template>
  <div class="d-flex flex-column w-100 position-relative bg-surface" style="border-radius: 12px; max-height: 40rem;">
    <div class="d-flex justify-space-between ga-6 pa-4 border-b-thin bg-surface align-center" style="z-index: 1; border-top-left-radius: 12px; border-top-right-radius: 12px;">
      <span class="text-title-large font-weight-bold">{{ $t('sources.asurascans.comic.chaptersCount', { count: chapters.length }) }}</span>
      <VBtn
        variant="tonal"
        class="text-body-medium"
        color="surfaceVariant"
        :text="sort === 'asc' ? $t('common.sort.oldest') : $t('common.sort.latest')"
        :prepend-icon="sort === 'desc' ? 'fa6-solid:arrow-down-short-wide' : 'fa6-solid:arrow-up-short-wide'"
        @click="sort = sort === 'asc' ? 'desc' : 'asc'"
      />
    </div>

    <div v-if="fetchChaptersLoading || filteredChapters.length === 0" class="d-flex justify-center align-center w-100 pa-4">
      <VProgressCircular
        v-if="fetchChaptersLoading"
        class="m-auto"
        indeterminate
        color="primary"
      />
      <span v-else class="text-medium-emphasis">{{ $t('sources.asurascans.comic.chaptersCount', { count: 0 }) }}</span>
    </div>
    <VVirtualScroll :items="filteredChapters" style="border-bottom-left-radius: 12px; border-bottom-right-radius: 12px;">
      <template #default="{ item }">
        <AsuraComicChaptersItem :chapter="item" />
      </template>
    </VVirtualScroll>
  </div>
</template>
