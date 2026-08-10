// SPDX-License-Identifier: AGPL-3.0-or-later

export interface OidcProviderStore {
  provider: Ref<OidcProviderDetails | undefined>
  setProvider: (value: OidcProviderDetails) => void

  providers: Ref<LightOidcProvider[]>
  setProviders: (value: LightOidcProvider[]) => void

  invalidate: () => void
}

export const useOidcProviderStore = defineStore('oidc-provider', (): OidcProviderStore => {
  const provider = ref<OidcProviderDetails>()

  function setProvider(value: OidcProviderDetails): void {
    provider.value = value
  }

  const providers = ref<LightOidcProvider[]>([])

  function setProviders(value: LightOidcProvider[]): void {
    providers.value = value
  }

  function invalidate(): void {
    providers.value = []
    provider.value = undefined
  }

  return {
    provider,
    providers,

    setProvider,
    setProviders,

    invalidate,
  }
})
