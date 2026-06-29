<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { Density } from 'vuetify/lib/composables/density.mjs'

const props = withDefaults(defineProps<{
  text: string
  color?: string
  size?: 'small' | 'default'
}>(), {
  size: 'default',
})

const { mobile } = useDisplay()

interface ChipProps {
  size?: string | number
  density?: Density
  class?: string
}

const chipProps = computed((): ChipProps => {
  if (props.size === 'small') {
    return {
      size: mobile.value ? 'x-small' : 'small',
      density: mobile.value ? 'comfortable' : 'compact',
      class: 'text-label-small',
    }
  }

  return {
    class: 'text-label-large',
  }
})
</script>

<template>
  <VChip
    :color
    variant="tonal"
    density="compact"
    v-bind="chipProps"
    rounded="pill"
    class="px-1"
    :text
  />
</template>
