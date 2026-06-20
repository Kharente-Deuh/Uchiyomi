<script setup lang="ts">
import { appName } from '~/constants'

import '~/assets/styles/reset.scss'
import '~/assets/styles/global.scss'
import '~/assets/styles/vuetify-overrides.scss'

useHead({
  titleTemplate: title => (title ? `${title} · ${appName}` : appName),
})
</script>

<template>
  <VitePwaManifest />
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
