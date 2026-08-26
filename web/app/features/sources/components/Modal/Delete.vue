<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { SourceComicInfos, SourceSearchItem } from '../../types'
import type { ComicSource } from '~/features/comics/types'

const props = defineProps<{
  source: ComicSource
}>()

const show = defineModel<boolean>({ required: true })
const comic = defineModel<SourceSearchItem | SourceComicInfos | undefined>('comic', { required: true })

const { removeComicFromLibrary } = useSourceSearch(props.source, { doSearch: false })

const loading = ref(false)

watch(show, (value) => {
  if (!value) {
    loading.value = false
    comic.value = undefined
  }
})

async function handleDelete(): Promise<void> {
  if (!comic.value) {
    return
  }

  loading.value = true

  await removeComicFromLibrary(comic.value as SourceSearchItem)

  show.value = false
}
</script>

<template>
  <OrganismModalConfirmation
    v-model="show"
    :loading
    :text="$t('sources.delete.confirmation', { name: comic?.title ?? '' })"
    @confirm="handleDelete"
  />
</template>
