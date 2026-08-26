<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceComicInfos } from '../../types'
import type { ComicSource } from '~/features/comics/types'

const props = defineProps<{
  source: ComicSource
  comicOriginUrl: string
}>()

const comic = defineModel<SourceComicInfos>({ required: true })

const { addComicInLibrary } = useSourceSearch(props.source, { doSearch: false })
const { fetchChapters } = useSourceChapters(props.source)

const showDeleteModal = ref(false)
const addToLibraryLoading = ref(false)

function toggleLibraryAction(): void {
  if (!comic.value) {
    return
  }

  if (comic.value.internalId) {
    showDeleteModal.value = true
  } else {
    addComicToLibrary()
  }
}

async function addComicToLibrary(): Promise<void> {
  if (!comic.value) {
    return
  }

  addToLibraryLoading.value = true

  const internalId = await addComicInLibrary(comic.value)
  if (internalId) {
    comic.value.internalId = internalId
  }

  addToLibraryLoading.value = false
}

watch(() => comic.value.internalId, () => {
  if (comic.value.slug) {
    fetchChapters(comic.value.slug)
  }
})
</script>

<template>
  <SourcesModalDelete
    v-if="comic"
    v-model="showDeleteModal"
    :source
    :comic
  />

  <div
    class="d-flex flex-column ga-4 pa-4 bg-surface"
    style="border-radius: 12px"
  >
    <div class="d-flex flex-wrap align-center justify-space-between ga-4">
      <div class="d-flex flex-wrap ga-4">
        <ComicsChipStatus :status="comic.status" />
        <ComicsChipType :type="comic.type" />
      </div>

      <SourcesComicLinkOriginalSite :to="`${props.comicOriginUrl}`" />
    </div>

    <VBtn
      variant="tonal"
      class="w-100"
      :color="comic.internalId ? 'error' : 'primary'"
      :class="{
        'border-thin-error': comic.internalId,
        'border-thin-primary': !comic.internalId,
      }"
      :prepend-icon="comic.internalId ? 'fa6-solid:trash' : 'fa6-solid:book'"
      size="large"
      :text="comic.internalId ? $t('comics.remove.title') : $t('sources.add.title')"
      :loading="addToLibraryLoading"
      @click="toggleLibraryAction"
    />
    <div v-if="comic.author" class="d-flex ga-2 justify-space-between">
      <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.author') }}</span>
      <span class="text-body-large font-weight-bold text-truncate">{{ comic.author }}</span>
    </div>
    <div v-if="comic.artist" class="d-flex ga-2 justify-space-between">
      <span class="text-body-large text-medium-emphasis text-uppercase text-truncate">{{ $t('comic.fields.artist') }}</span>
      <span class="text-body-large font-weight-bold text-truncate">{{ comic.artist }}</span>
    </div>
  </div>
</template>
