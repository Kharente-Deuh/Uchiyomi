<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicStatus } from '../../types'

const props = defineProps<{ status: ComicStatus }>()

const borderClass = computed(() => {
  switch (props.status) {
    case 'ongoing':
      return 'border-thin-primary'
    case 'completed':
      return 'border-thin'
    case 'hiatus':
      return 'border-thin-warning'
    case 'cancelled':
      return 'border-thin-error'
    default:
      return 'border-thin'
  }
})
</script>

<template>
  <div
    class="d-flex align-center ga-3 w-fit px-2 py-1 bg-background"
    :style="{ borderRadius: '12px' }"
    :class="borderClass"
  >
    <ComicsIconStatus :status />
    <span class="text-body-large text-truncate">{{ $t(`sources.asura.status.${status}`) }}</span>
  </div>
</template>
