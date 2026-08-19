<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Comic } from '../types'

const props = defineProps<{ comic: Comic }>()

const showMore = ref(false)

const sanitzedDescription = computed(() => props.comic.description.replace(/<[^>]*>?/g, '').replaceAll('\n', ' '))
</script>

<template>
  <div class="d-flex flex-column ga-3 transition-smooth pa-4 bg-surface w-100" style="border-radius: 12px;">
    <span
      v-if="sanitzedDescription"
      class="text-body-large transition-smooth"
      :class="{ 'text-truncate': !showMore }"
    >{{ sanitzedDescription }}</span>

    <span
      v-if="comic.altTitles"
      class="transition-smooth text-medium-emphasis text-body-small"
      :class="{ 'text-truncate': !showMore }"
    >{{ comic.altTitles.join(', ') }}</span>

    <div v-if="comic.genres.length && showMore" class="d-flex flex-wrap ga-3">
      <span
        v-for="(genre, i) in comic.genres"
        :key="i"
        class="text-capitalize px-2 py-1 bg-background border-thin text-body-medium rounded-lg"
      >{{ genre }}</span>
    </div>

    <div class="d-flex justify-end">
      <VBtn
        variant="text"
        color="primary"
        :text="showMore ? $t('common.showLess') : $t('common.showMore')"
        :append-icon="showMore ? 'fa6-solid:angle-up' : 'fa6-solid:angle-down'"
        @click="showMore = !showMore"
      />
    </div>
  </div>
</template>
