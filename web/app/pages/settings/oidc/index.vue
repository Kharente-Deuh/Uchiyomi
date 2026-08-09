<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP, ADMIN_ROUTE_GROUP],
})

const { providers, getAll } = useOidc()

const showCreateModal = ref(false)

const globalLoading = ref(false)
async function init(): Promise<void> {
  globalLoading.value = true

  await getAll()

  globalLoading.value = false
}

onMounted(() => {
  init()
})
</script>

<template>
  <OrganismPageLayout
    :title="$t('settings.oidc.title')"
    icon="fa6-solid:gear"
    :loading="globalLoading"
  >
    <template #append-title>
      <VBtn
        v-tooltip="$t('settings.oidc.create.title')"
        :disabled="globalLoading"
        color="primary"
        variant="tonal"
        class="border-thin-primary"
        icon="fa6-solid:plus"
        density="comfortable"
        @click="showCreateModal = true"
      />
    </template>

    <OidcModalCreate v-if="!globalLoading" v-model="showCreateModal" />

    <div v-if="providers.length === 0 && !globalLoading" class="d-flex flex-column ga-4">
      <span class="mx-auto">{{ $t('settings.oidc.no-providers') }}</span>
      <VBtn
        color="primary"
        variant="tonal"
        style="width: fit-content"
        prepend-icon="fa6-solid:plus"
        class="border-thin-primary mx-auto"
        :text="$t('settings.oidc.create.title')"
        @click="showCreateModal = true"
      />
    </div>

    <div v-if="providers.length > 0" class="providers-grid">
      <OidcCardProvider
        v-for="(p, i) of providers"
        :key="i"
        v-bind="p"
      />
    </div>
  </OrganismPageLayout>
</template>

<style lang="scss">
.providers-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
}
</style>
