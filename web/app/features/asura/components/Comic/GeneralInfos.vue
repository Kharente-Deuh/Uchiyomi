<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraComicInfos } from '../../types'

const props = defineProps<{
  comic: AsuraComicInfos
}>()

const showMore = ref(false)

const sanitzedDescription = computed(() => props.comic.description.replace(/<[^>]*>?/g, '').replaceAll('\n', ' '))
</script>

<template>
  <div class="d-flex flex-column ga-6 transition-smooth pa-4 bg-surface w-100" style="border-radius: 12px;">
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
