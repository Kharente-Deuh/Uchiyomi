// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { OidcProvider } from './oidc.api'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useOidcProvider } from './oidc-provider.composable'

const { getById, updateById, testByIssuerUrl, deleteById, navigateTo } = vi.hoisted(() => ({
  getById: vi.fn(),
  updateById: vi.fn(),
  testByIssuerUrl: vi.fn(),
  deleteById: vi.fn(),
  navigateTo: vi.fn(),
}))

vi.mock('./oidc.api', () => ({
  createOidcApi: () => ({ getById, updateById, testByIssuerUrl, deleteById }),
}))

vi.mock('vue-router', async importOriginal => ({
  ...(await importOriginal<typeof import('vue-router')>()),
  onBeforeRouteLeave: vi.fn(),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

mockNuxtImport('useI18n', () => i18nStub)
mockNuxtImport('navigateTo', () => navigateTo)

const provider = {
  id: 'p1',
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  usernameClaim: 'preferred_username',
  scopes: ['openid'],
  autoProvision: false,
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  updatedAt: new Date('2026-02-03T04:05:06.000Z'),
} satisfies OidcProvider

function apiError(status: number): { success: false, error: { status: number } } {
  return { success: false, error: { status } }
}

beforeEach(() => {
  setActivePinia(createPinia())
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getById.mockReset()
  updateById.mockReset()
  testByIssuerUrl.mockReset()
  deleteById.mockReset()
  navigateTo.mockReset()
})

async function withProvider(): Promise<ReturnType<typeof useOidcProvider>> {
  getById.mockResolvedValue({ success: true, data: provider })
  const composable = useOidcProvider()
  await composable.fetchProvider('p1')
  useToast().messages.value.length = 0

  return composable
}

describe('useOidcProvider().fetchProvider', () => {
  it('exposes the fetched provider', async () => {
    getById.mockResolvedValue({ success: true, data: provider })
    const composable = useOidcProvider()

    await composable.fetchProvider('p1')

    expect(composable.provider.value).toEqual(provider)
    expect(getById).toHaveBeenCalledWith('p1')
  })

  it('leaves fetchLoading false once the provider is in', async () => {
    getById.mockResolvedValue({ success: true, data: provider })
    const composable = useOidcProvider()

    await composable.fetchProvider('p1')

    expect(composable.fetchLoading.value).toBe(false)
  })

  it('toasts a not-found message and leaves the page on a 404', async () => {
    getById.mockResolvedValue(apiError(404))

    await useOidcProvider().fetchProvider('nope')

    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.errors.notFound', color: 'error' }])
    expect(navigateTo).toHaveBeenCalledWith('/settings/oidc')
  })

  it('toasts an unknown error and leaves the page on any other failure', async () => {
    getById.mockResolvedValue(apiError(500))

    await useOidcProvider().fetchProvider('p1')

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
    expect(navigateTo).toHaveBeenCalledWith('/settings/oidc')
  })

  it('does not navigate away on success', async () => {
    getById.mockResolvedValue({ success: true, data: provider })

    await useOidcProvider().fetchProvider('p1')

    expect(navigateTo).not.toHaveBeenCalled()
  })
})

describe('useOidcProvider().update', () => {
  it('strips id and timestamps from the payload', async () => {
    updateById.mockResolvedValue({ success: true, data: provider })

    await useOidcProvider().update({ ...provider, displayName: 'Renamed' })

    expect(updateById).toHaveBeenCalledWith('p1', {
      displayName: 'Renamed',
      issuerUrl: provider.issuerUrl,
      clientId: provider.clientId,
      usernameClaim: provider.usernameClaim,
      scopes: provider.scopes,
      autoProvision: provider.autoProvision,
    })
  })

  it('refreshes the held provider with the server response', async () => {
    updateById.mockResolvedValue({ success: true, data: { ...provider, displayName: 'Renamed' } })
    const composable = useOidcProvider()

    await composable.update(provider)

    expect(composable.provider.value?.displayName).toBe('Renamed')
    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.update.success', color: 'success' }])
  })

  it('toasts a not-found message on a 404 without navigating away', async () => {
    updateById.mockResolvedValue(apiError(404))

    await useOidcProvider().update(provider)

    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.errors.notFound', color: 'error' }])
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('toasts an unknown error on any other failure', async () => {
    updateById.mockResolvedValue(apiError(500))

    await useOidcProvider().update(provider)

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })

  it('leaves updateLoading false once settled', async () => {
    updateById.mockResolvedValue(apiError(500))
    const composable = useOidcProvider()

    await composable.update(provider)

    expect(composable.updateLoading.value).toBe(false)
  })
})

describe('useOidcProvider().test', () => {
  it('does nothing while no provider is held', async () => {
    await expect(useOidcProvider().test()).resolves.toBeUndefined()
    expect(testByIssuerUrl).not.toHaveBeenCalled()
  })

  it('probes the issuer url of the held provider', async () => {
    getById.mockResolvedValue({ success: true, data: provider })
    testByIssuerUrl.mockResolvedValue({ success: true, data: { issuer: provider.issuerUrl } })
    const composable = useOidcProvider()

    await composable.fetchProvider('p1')

    await expect(composable.test()).resolves.toEqual({ issuer: provider.issuerUrl })
    expect(testByIssuerUrl).toHaveBeenCalledWith(provider.issuerUrl)
  })

  it('leaves testLoading false once the probe settled', async () => {
    testByIssuerUrl.mockResolvedValue({ success: true, data: { issuer: provider.issuerUrl } })
    const composable = await withProvider()

    await composable.test()

    expect(composable.testLoading.value).toBe(false)
  })
})

describe('useOidcProvider().deleteProvider', () => {
  it('does nothing while no provider is held', async () => {
    await useOidcProvider().deleteProvider()

    expect(deleteById).not.toHaveBeenCalled()
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('drops the held provider and leaves the page on success', async () => {
    deleteById.mockResolvedValue({ success: true, data: undefined })
    const composable = await withProvider()

    await composable.deleteProvider()

    expect(deleteById).toHaveBeenCalledWith('p1')
    expect(composable.provider.value).toBeUndefined()
    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.delete.success', color: 'success' }])
    expect(navigateTo).toHaveBeenCalledWith('/settings/oidc')
  })

  it('treats a 404 as a successful deletion', async () => {
    deleteById.mockResolvedValue(apiError(404))
    const composable = await withProvider()

    await composable.deleteProvider()

    expect(composable.provider.value).toBeUndefined()
    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.delete.success', color: 'success' }])
    expect(navigateTo).toHaveBeenCalledWith('/settings/oidc')
  })

  it('keeps the provider and stays on the page on any other failure', async () => {
    deleteById.mockResolvedValue(apiError(500))
    const composable = await withProvider()

    await composable.deleteProvider()

    expect(composable.provider.value).toEqual(provider)
    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
    expect(navigateTo).not.toHaveBeenCalled()
  })

  it('leaves deleteLoading false once settled', async () => {
    deleteById.mockResolvedValue(apiError(500))
    const composable = await withProvider()

    await composable.deleteProvider()

    expect(composable.deleteLoading.value).toBe(false)
  })
})

describe('useOidcProvider().invalidate', () => {
  it('drops the held provider', async () => {
    const composable = await withProvider()

    composable.invalidate()

    expect(composable.provider.value).toBeUndefined()
  })
})
