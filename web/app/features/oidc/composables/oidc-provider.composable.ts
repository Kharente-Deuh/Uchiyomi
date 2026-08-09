// SPDX-License-Identifier: AGPL-3.0-or-later

export interface OidcProviderComposable {
  provider: Ref<OidcProvider | undefined>

  fetchProvider: (id: string) => Promise<void>
  fetchLoading: Ref<boolean>

  update: (p: OidcProvider) => Promise<void>
  updateLoading: Ref<boolean>

  test: () => Promise<TestResponse | undefined>
  testLoading: Ref<boolean>
}

export function useOidcProvider(): OidcProviderComposable {
  const { t } = useI18n()
  const toast = useToast()
  const api = createOidcApi()
  const composable = useOidc()
  const store = useOidcProviderStore()

  const provider = computed(() => store.provider)

  const fetchLoading = ref(false)
  const updateLoading = ref(false)

  async function fetchProvider(id: string): Promise<void> {
    fetchLoading.value = true
    const res = await api.getById(id)
    if (res.success) {
      store.setProvider(res.data)

      fetchLoading.value = false

      return
    }

    switch (res.error.status) {
      case 404:
        toast.error(t('settings.oidc.errors.notFound'))
        break
      default:
        toast.error(t('error.unknown'))
    }

    navigateTo('/settings/oidc')
  }

  async function update(p: OidcProvider): Promise<void> {
    updateLoading.value = true
    const { id, createdAt: _createdAt, updatedAt: _updatedAt, ...data } = p
    const res = await api.updateById(id, data)
    updateLoading.value = false

    if (res.success) {
      toast.success(t('settings.oidc.update.success'))
      store.setProvider(res.data)

      return
    }

    switch (res.error.status) {
      case 404:
        toast.error(t('settings.oidc.errors.notFound'))
        break
      default:
        toast.error(t('error.unknown'))
    }
  }

  async function test(): Promise<TestResponse | undefined> {
    if (!provider.value) {
      return
    }

    return composable.test(provider.value.issuerUrl)
  }

  onBeforeRouteLeave(() => {
    store.invalidate()
  })

  return {
    provider,

    fetchProvider,
    fetchLoading,

    update,
    updateLoading,

    test,
    testLoading: composable.testLoading,
  }
}
