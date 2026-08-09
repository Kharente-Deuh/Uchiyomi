<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { ADMIN_ROUTE_GROUP, AUTHENTICATED_ROUTE_GROUP } from '~/constants/auth'

definePageMeta({
  layout: 'default',
  authGroups: [AUTHENTICATED_ROUTE_GROUP, ADMIN_ROUTE_GROUP],
})

const route = useRoute('settings-oidc-id')
const { smAndDown } = useDisplay()
const { fetchProvider, fetchLoading, provider } = useOidcProvider()

onMounted(() => {
  fetchProvider(route.params.id)
})
</script>

<template>
  <OrganismPageLayout
    :title="provider?.displayName ?? $t('common.loading')"
    back-route="/settings/oidc"
    :show-back-route="true"
    :global-loader="fetchLoading"
    :back-route-name="$t('settings.oidc.titleShort')"
  >
    <div class="d-flex flex-column ga-6" :class="{ 'px-6': smAndDown }">
      <OidcCardCategoryInformations :loading="fetchLoading" />
      <OidcCardCategoryClaims :loading="fetchLoading" />
    </div>
  </OrganismPageLayout>
</template>
