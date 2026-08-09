<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PageLayoutBackRoute } from '~/components/Organism/PageLayout.vue'
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP, ADMIN_ROUTE_GROUP],
})

const route = useRoute('settings-oidc-id')
const { t } = useI18n()
const { smAndDown } = useDisplay()
const { fetchProvider, fetchLoading, provider } = useOidcProvider()

onMounted(() => {
  fetchProvider(route.params.id)
})

const backRoutes = computed((): PageLayoutBackRoute[] => [
  ...(smAndDown.value ? [{ to: '/settings', name: t('settings.title') }] : []),
  {
    to: '/settings/oidc',
    name: t('settings.oidc.titleShort'),
  },
])
</script>

<template>
  <OrganismPageLayout
    :title="provider?.displayName ?? $t('common.loading')"
    :back-routes
    :global-loader="fetchLoading"
  >
    <div class="d-flex flex-column ga-6" :class="{ 'px-6': smAndDown }">
      <OidcCardCategoryInformations :loading="fetchLoading" />
      <OidcCardCategoryClaims :loading="fetchLoading" />
    </div>
  </OrganismPageLayout>
</template>
