<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
defineProps<LightOidcProvider>()

const { locale } = useI18n()
</script>

<template>
  <AtomLink :to="`/settings/oidc/${id}`">
    <VCard
      class="provider-card pa-4 border-thin"
      style="border-radius: 12px;"
    >
      <div class="d-flex ga-4 flex-column">
        <div class="d-flex align-center text-truncate ga-4 justify-space-between">
          <span class="text-truncate text-title-large">{{ displayName }}</span>
          <VIcon icon="fa6-solid:arrow-up-right-from-square" />
        </div>
        <div class="d-flex justify-space-between ga-4 align-center">
          <div class="d-flex align-center ga-2 text-body-medium">
            <span> {{ userCount }} </span>
            <VIcon icon="fa6-solid:users" />
          </div>

          <span class="text-body-medium">{{ createdAt.toLocaleDateString(locale) }}</span>
        </div>
      </div>
    </VCard>
  </AtomLink>
</template>
