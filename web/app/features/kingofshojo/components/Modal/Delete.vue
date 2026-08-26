<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { KingOfShojoComicInfos, KingOfShojoSearchItem } from '../../types'

const show = defineModel<boolean>({ required: true })
const comic = defineModel<KingOfShojoSearchItem | KingOfShojoComicInfos | undefined>('comic', { required: true })

const { removeComicFromLibrary } = useKingOfShojoSearch({ doSearch: false })

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

  await removeComicFromLibrary(comic.value as KingOfShojoSearchItem)

  show.value = false
}
</script>

<template>
  <OrganismModalConfirmation
    v-model="show"
    :loading
    :text="$t('sources.kingofshojo.delete.confirmation', { name: comic?.title ?? '' })"
    @confirm="handleDelete"
  />
</template>
