<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { APP_NAME } from '~/constants'

import '~/assets/styles/reset.scss'
import '~/assets/styles/global.scss'
import '~/assets/styles/vuetify-overrides.scss'
import '~/assets/styles/constants.scss'

useHead({
  titleTemplate: title => (title ? `${title} · ${APP_NAME}` : APP_NAME),
})

function pageKey(route: RouteLocationNormalizedLoaded): string {
  if (route.name === 'comic-id-chapterId' && 'id' in route.params) {
    return `comic-${route.params.id}`
  }

  return route.fullPath
}
</script>

<template>
  <VitePwaManifest />
  <NuxtLoadingIndicator color="primary" :height="2" />
  <NuxtLayout>
    <NuxtPage :page-key="pageKey" />
  </NuxtLayout>
</template>
