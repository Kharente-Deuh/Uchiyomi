<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../types'

defineEmits<{ deleted: [] }>()
const comic = defineModel<Comic>({ required: true })
const { t } = useI18n()
const toast = useToast()
const api = createComicsApi()

const showDeleteModal = ref(false)
const refreshLoading = ref(false)

async function refreshComic(): Promise<void> {
  refreshLoading.value = true

  const response = await api.refreshById(comic.value.id)
  if (response.success) {
    comic.value = response.data
  } else {
    console.error(response.error)
    toast.error(t('error.unknown'))
  }

  refreshLoading.value = false
}

const canRefresh = computed(() => comic.value.status !== 'hiatus' && comic.value.status !== 'completed' && comic.value.status !== 'cancelled')
</script>

<template>
  <ComicsModalDelete
    v-if="comic"
    v-model="showDeleteModal"
    :comic
    @deleted="$emit('deleted')"
  />

  <div
    class="d-flex flex-column ga-4 pa-4 bg-surface"
    style="border-radius: 12px"
  >
    <div class="d-flex flex-wrap ga-4">
      <ComicsChipStatus :status="comic.status" />
      <ComicsChipType :type="comic.type" />
    </div>

    <div class="d-flex ga-2 ga-4 align-center flex-wrap">
      <VBtn
        v-if="canRefresh"
        v-tooltip="$t('comics.refresh.title')"
        icon="fa6-solid:repeat"
        color="secondary"
        size="small"
        class="border-thin-secondary"
        @click="refreshComic"
      />

      <VBtn
        variant="tonal"
        class="border-thin-error"
        color="error"
        prepend-icon="fa6-solid:trash"
        size="large"
        :text="$t('comics.remove.title')"
        @click="showDeleteModal = true"
      />
    </div>
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
