<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import asuraImg from '~/assets/images/sources/asurascans.webp'
import { AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP],
})
</script>

<template>
  <OrganismPageLayout :title="`${$t('sources.asurascans.title')} yolo`">
    <template #prepend-title>
      <img
        aspect-ratio="1"
        :src="asuraImg"
        class="border-thin rounded-lg"
        style="height: 48px;"
      >
    </template>
  </OrganismPageLayout>
</template>
