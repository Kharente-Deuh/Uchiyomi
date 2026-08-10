<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { ProviderSummary } from '~/features/auth/composables/auth.api'
import { oidcStartUrl } from '~/features/auth/composables/auth.api'

interface Props {
  providers: ProviderSummary[]
  redirect: string
}

defineProps<Props>()
const { t } = useI18n()
</script>

<template>
  <div v-if="providers.length" class="d-flex flex-column ga-2 px-4">
    <VBtn
      v-for="provider in providers"
      :key="provider.id"
      :href="oidcStartUrl(provider.id, redirect)"
      variant="outlined"
      size="large"
      class="w-100"
      :data-test="`login-oidc-${provider.id}`"
      :text="t('login.oidc.button', { name: provider.displayName })"
    />
  </div>
</template>
