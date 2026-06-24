<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import logoHorizontalDark from '~/assets/images/logo-horizontal-dark.webp'
import logoHorizontalLight from '~/assets/images/logo-horizontal.webp'

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)
</script>

<template>
  <VApp>
    <VMain class="bg-bimo-grey-darken-2 d-flex flex-column justify-center h-100 gradient-background">
      <div
        style="width: 23rem;"
        class="ma-auto d-flex flex-column ga-6 justify-center"
      >
        <VImg
          :src="isDark ? logoHorizontalDark : logoHorizontalLight"
          class="w-100"
        />

        <slot />
      </div>
    </VMain>
  </VApp>
</template>

<style lang="scss" scoped>
.auth-card {
  margin-inline: auto;
}
</style>
