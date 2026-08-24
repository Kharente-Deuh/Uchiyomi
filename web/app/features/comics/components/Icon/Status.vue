<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ComicStatus } from '~/features/comics/types'

const props = defineProps<{
  status: ComicStatus
  withBackground?: boolean
}>()

const modelValue = computed((): { icon: string, color: string } | undefined => {
  switch (props.status) {
    case 'ongoing':
      return { icon: 'fa6-regular:circle-dot', color: 'primary' }
    case 'completed':
      return { icon: 'fa6-solid:check', color: 'grey' }
    case 'hiatus':
      return { icon: 'fa6-solid:pause', color: 'warning' }
    case 'cancelled':
      return { icon: 'fa6-solid:ban', color: 'error' }
    default:
      return undefined
  }
})
</script>

<template>
  <template v-if="modelValue">
    <VIcon
      v-if="!withBackground"
      :icon="modelValue.icon"
      :color="modelValue.color"
    />

    <div v-else class="d-flex flex-column items-center justify-center rounded-lg border-thin transition-smooth status-icon-box">
      <VIcon
        v-tooltip="$t(`sources.asurascans.status.${status}`)"
        :icon="modelValue.icon"
        :color="modelValue.color"
        size="x-small"
      />
    </div>
  </template>
</template>

<style lang="scss" scoped>
.status-icon-box {
  padding: 0.4rem;
  background-color: rgba(var(--v-theme-surface), 0.7);
  backdrop-filter: blur(2px);

  &:hover {
    background-color: rgb(var(--v-theme-surface));
  }
}
</style>
