<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicType } from '../../types'

withDefaults(defineProps<{
  type: ComicType
  size?: 'default' | 'small'
}>(), { size: 'default' })
</script>

<template>
  <span
    class="ga-3 w-fit h-fit px-2 py-1 border-thin text-truncate text-uppercase bg-background"
    :class="{
      'text-body-large': size === 'default',
      'text-body-medium': size === 'small',
    }"
    :style="{ borderRadius: '12px' }"
  >
    <span class="">{{ $t(`sources.asurascans.type.${type}`) }}</span>
  </span>
</template>
