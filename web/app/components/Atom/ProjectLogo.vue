<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import LogoHorizontalDark from '~/assets/images/logo-horizontal-dark.webp'
import LogoHorizontalLight from '~/assets/images/logo-horizontal.webp'
import LogoSquare from '~/assets/images/logo-square.webp'

const props = defineProps<{ compact?: boolean }>()

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)

const src = computed(() => {
  if (props.compact) {
    return LogoSquare
  }

  return isDark.value ? LogoHorizontalDark : LogoHorizontalLight
})

const aspectRatio = computed(() => {
  if (!props.compact) {
    return
  }

  return '1/1'
})
</script>

<template>
  <VImg
    class="transition-smooth w-100"
    :src="src"
    :aspect-ratio="aspectRatio"
  />
</template>
