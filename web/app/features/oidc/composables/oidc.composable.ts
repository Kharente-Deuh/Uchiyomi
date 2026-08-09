// SPDX-License-Identifier: AGPL-3.0-or-later

import type { CreateOidcProviderRequest, LightOidcProvider, TestResponse } from './oidc.api'
import { createOidcApi } from './oidc.api'

export interface OidcComposable {
  providers: Ref<LightOidcProvider[]>
  loading: Ref<boolean>
  testLoading: Ref<boolean>

  getAll: () => Promise<void>
  create: (request: CreateOidcProviderRequest) => Promise<void>
  test: (issuerUrl: string) => Promise<TestResponse | undefined>
  deleteById: (id: string) => Promise<void>
}

export function useOidc(): OidcComposable {
  const { t } = useI18n()
  const toast = useToast()
  const api = createOidcApi()

  const providers = ref<LightOidcProvider[]>([])
  const loading = ref(false)
  const testLoading = ref(false)

  async function getAll(): Promise<void> {
    loading.value = true
    const res = await api.getAll()
    if (res.success) {
      providers.value = res.data
    } else {
      console.error(res.error)
      toast.error(t('error.unknown'))
    }

    loading.value = false
  }

  async function create(request: CreateOidcProviderRequest): Promise<void> {
    loading.value = true
    const res = await api.create(request)
    if (res.success) {
      providers.value = [
        {
          id: res.data.id,
          displayName: res.data.displayName,
          createdAt: res.data.createdAt,
          userCount: 0,
        },
        ...providers.value,
      ]

      toast.success(t('settings.oidc.create.success', { name: res.data.displayName }))
    } else {
      switch (res.error.status) {
        case 409:
          toast.error(t('settings.oidc.create.error'))

          break
        default:
          toast.error(t('error.unknown'))
      }
    }

    loading.value = false
  }

  async function test(issuerUrl: string): Promise<TestResponse | undefined> {
    testLoading.value = true
    const res = await api.testByIssuerUrl(issuerUrl)
    if (res.success) {
      testLoading.value = false

      return res.data
    }

    if (res.error.status === 400 && res.error.message === 'issuer is unreachable') {
      toast.error(t('settings.oidc.test.error.unreachable'))

      testLoading.value = false

      return
    }

    console.error('api.testIssuerByUrl', res.error)
    toast.error(t('error.unknown'))

    testLoading.value = false
  }

  async function deleteById(id: string): Promise<void> {
    loading.value = true
    const res = await api.deleteById(id)
    if (res.success || res.error.status === 404) {
      toast.success(t('settings.oidc.delete.success'))
      providers.value = providers.value.filter(p => p.id !== id)
    } else {
      console.error('api.deleteById', res.error)
      toast.error(t('error.unknown'))
    }

    loading.value = false
  }

  return {
    providers,
    loading,
    testLoading,

    getAll,
    create,
    test,
    deleteById,
  }
}
