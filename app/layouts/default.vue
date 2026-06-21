<script setup lang="ts">
// SPDX-License-Identifier: AGPL-3.0-or-later
import { useDisplay } from 'vuetify'

const display = useDisplay()
const route = useRoute()
const { items } = useNavigation()
const { showOverlay } = useOrientationLock()

function go(to: string): void {
  if (to !== route.path) {
    navigateTo(to)
  }
}
</script>

<template>
  <VApp>
    <AppBarChrome>
      <LanguageSwitch />
      <DarkToggle />
    </AppBarChrome>

    <AppNavigation
      v-if="!display.mobile.value"
      variant="rail"
      :items="items"
      :active="route.path"
      @navigate="go"
    />

    <VMain>
      <VContainer>
        <slot />
      </VContainer>
    </VMain>

    <AppNavigation
      v-if="display.mobile.value"
      variant="bottom"
      :items="items"
      :active="route.path"
      @navigate="go"
    />

    <OrientationOverlay :model-value="showOverlay" />
  </VApp>
</template>
