<script setup lang="ts">
import { appName } from '~/constants'

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
