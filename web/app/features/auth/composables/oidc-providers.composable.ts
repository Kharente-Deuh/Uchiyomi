// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ProviderSummary } from './auth.api'
import { createAuthApi } from './auth.api'

export interface OIDCProvidersComposable {
  providers: Ref<ProviderSummary[]>
  loading: Ref<boolean>
  fetchProviders: () => Promise<void>
}

export function useOIDCProviders(): OIDCProvidersComposable {
  const api = createAuthApi()
  const providers = ref<ProviderSummary[]>([])
  const loading = ref(false)

  async function fetchProviders(): Promise<void> {
    loading.value = true

    const res = await api.getProviders()
    providers.value = res.success ? res.data : []
    if (!res.success) {
      console.error('useOIDCProviders.fetchProviders', res.error)
    }

    loading.value = false
  }

  return { providers, loading, fetchProviders }
}
