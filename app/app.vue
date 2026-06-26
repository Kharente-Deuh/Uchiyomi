<script setup lang="ts">
import { appName } from '~/constants'

import '~/assets/styles/reset.scss'
import '~/assets/styles/global.scss'
import '~/assets/styles/vuetify-overrides.scss'
import '~/assets/styles/constants.scss'

useHead({
  titleTemplate: title => (title ? `${title} · ${appName}` : appName),
})
</script>

<template>
  <VitePwaManifest />
  <NuxtLoadingIndicator />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
