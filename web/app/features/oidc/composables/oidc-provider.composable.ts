// SPDX-License-Identifier: AGPL-3.0-or-later

export interface OidcProviderComposable {
  provider: Ref<OidcProviderDetails | undefined>

  fetchProvider: (id: string) => Promise<void>
  fetchLoading: Ref<boolean>

  update: (p: OidcProvider) => Promise<void>
  updateLoading: Ref<boolean>

  test: () => Promise<TestResponse | undefined>
  testLoading: Ref<boolean>

  deleteProvider: () => Promise<void>
  deleteLoading: Ref<boolean>

  invalidate: () => void
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
  const deleteLoading = ref(false)

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
    if (!provider.value) {
      return
    }

    updateLoading.value = true
    const { id, createdAt: _createdAt, updatedAt: _updatedAt, ...data } = p
    const res = await api.updateById(id, data)
    updateLoading.value = false

    if (res.success) {
      toast.success(t('settings.oidc.update.success'))
      store.setProvider({ ...res.data, users: provider.value.users })

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

  async function deleteProvider(): Promise<void> {
    if (!provider.value) {
      return
    }

    deleteLoading.value = true
    const res = await api.deleteById(provider.value.id)
    if (res.success || res.error.status === 404) {
      toast.success(t('settings.oidc.delete.success'))
      store.invalidate()
      await navigateTo('/settings/oidc')
    } else {
      console.error('api.deleteById', res.error)
      toast.error(t('error.unknown'))
    }

    deleteLoading.value = false
  }

  return {
    provider,

    fetchProvider,
    fetchLoading,

    update,
    updateLoading,

    test,
    testLoading: composable.testLoading,

    deleteProvider,
    deleteLoading,

    invalidate: store.invalidate,
  }
}
