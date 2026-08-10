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
const { fetchProvider, fetchLoading, provider, invalidate } = useOidcProvider()

const showDeleteModal = ref(false)

onMounted(() => {
  fetchProvider(route.params.id)
})

onBeforeRouteLeave(invalidate)

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
    :title="provider?.displayName ?? ''"
    :back-routes
    :loading="fetchLoading"
    global-loader
  >
    <div class="d-flex flex-column ga-6" :class="{ 'px-6': smAndDown }">
      <OidcCardCategoryInformations :loading="fetchLoading" />
      <OidcCardCategoryClaims :loading="fetchLoading" />
      <OidcCardCategoryUsers :loading="fetchLoading" />
    </div>

    <OidcModalDelete v-if="provider" v-model="showDeleteModal" />

    <div class="d-flex justify-center w-100 pt-6" :class="{ 'px-6': smAndDown }">
      <VBtn
        :text="$t('actions.delete')"
        color="error"
        prepend-icon="fa6-regular:trash-can"
        class="border-thin-error"
        :disabled="!provider"
        @click="showDeleteModal = true"
      />
    </div>
  </OrganismPageLayout>
</template>
