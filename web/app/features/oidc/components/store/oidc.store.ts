// SPDX-License-Identifier: AGPL-3.0-or-later

export interface OidcProviderStore {
  provider: Ref<OidcProvider | undefined>
  setProvider: (value: OidcProvider) => void
  invalidate: () => void
}

export const useOidcProviderStore = defineStore('oidc-provider', (): OidcProviderStore => {
  const provider = ref<OidcProvider>()

  function setProvider(value: OidcProvider): void {
    provider.value = value
  }

  function invalidate(): void {
    provider.value = undefined
  }

  return {
    provider,

    setProvider,
    invalidate,
  }
})
