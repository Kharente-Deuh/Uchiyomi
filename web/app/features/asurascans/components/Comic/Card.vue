<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { AsuraScansSearchItem } from '../../types'
import { ASURA_SOURCE_NAME } from '~/constants'

defineProps<{
  comic: AsuraScansSearchItem
  loading?: boolean
}>()

defineEmits<{ toggle: [] }>()
</script>

<template>
  <SourcesCardComic
    :internal-id="comic.internalId"
    :status="comic.status"
    :to="`/browse/sources/${ASURA_SOURCE_NAME}/${comic.slug}`"
    :cover="comic.cover"
    :title="comic.title"
    :chapter-count="comic.chapterCount"
    :loading
    @toggle="$emit('toggle')"
  />
</template>
