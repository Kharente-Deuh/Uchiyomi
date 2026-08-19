<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../types'

defineProps<{ comic: Comic }>()
defineEmits<{ deleted: [] }>()
const showDeleteModal = ref(false)
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

    <VBtn
      variant="tonal"
      class="w-100 border-thin-error"
      color="error"
      prepend-icon="fa6-solid:trash"
      size="large"
      :text="$t('comics.remove.title')"
      @click="showDeleteModal = true"
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
