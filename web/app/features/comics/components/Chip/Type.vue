<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicType } from '../../types'

defineProps<{ type: ComicType }>()
</script>

<template>
  <span
    class="ga-3 w-fit px-2 py-1 border-thin text-body-large text-truncate text-uppercase bg-background"
    :style="{ borderRadius: '12px' }"
  >
    <span class="">{{ $t(`sources.asura.type.${type}`) }}</span>
  </span>
</template>
