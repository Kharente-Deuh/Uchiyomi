// SPDX-License-Identifier: AGPL-3.0-or-later
// @vitest-environment nuxt

import type { LightOidcProvider } from './oidc.api'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useToast } from '~/composables/toast.composable'
import { useOidc } from './oidc.composable'

const { getAll, create, testByIssuerUrl, deleteById } = vi.hoisted(() => ({
  getAll: vi.fn(),
  create: vi.fn(),
  testByIssuerUrl: vi.fn(),
  deleteById: vi.fn(),
}))

vi.mock('./oidc.api', () => ({
  createOidcApi: () => ({ getAll, create, testByIssuerUrl, deleteById }),
}))

function i18nStub(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

mockNuxtImport('useI18n', () => i18nStub)

const provider = {
  id: 'p1',
  displayName: 'PocketID',
  createdAt: new Date('2026-01-02T03:04:05.000Z'),
  userCount: 0,
} satisfies LightOidcProvider

const request = {
  displayName: 'PocketID',
  issuerUrl: 'https://id.example.org',
  clientId: 'client',
  clientSecret: 'secret',
  usernameClaim: 'preferred_username',
  scopes: ['openid'],
  autoProvision: false,
}

function apiError(status: number, message = 'boom'): { success: false, error: { status: number, message: string } } {
  return { success: false, error: { status, message } }
}

beforeEach(() => {
  useToast().messages.value.length = 0
  vi.spyOn(console, 'error').mockImplementation(() => {})
  getAll.mockReset()
  create.mockReset()
  testByIssuerUrl.mockReset()
  deleteById.mockReset()
})

describe('useOidc().getAll', () => {
  it('fills the provider list', async () => {
    getAll.mockResolvedValue({ success: true, data: [provider] })
    const oidc = useOidc()

    await oidc.getAll()

    expect(oidc.providers.value).toEqual([provider])
  })

  it('leaves loading false once settled', async () => {
    getAll.mockResolvedValue({ success: true, data: [] })
    const oidc = useOidc()

    await oidc.getAll()

    expect(oidc.loading.value).toBe(false)
  })

  it('toasts an unknown error on failure', async () => {
    getAll.mockResolvedValue(apiError(500))
    const oidc = useOidc()

    await oidc.getAll()

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
    expect(oidc.providers.value).toEqual([])
  })
})

describe('useOidc().create', () => {
  it('prepends the created provider with a zero user count', async () => {
    getAll.mockResolvedValue({ success: true, data: [{ ...provider, id: 'p0', displayName: 'Keycloak' }] })
    create.mockResolvedValue({ success: true, data: { ...provider, userCount: 99 } })
    const oidc = useOidc()

    await oidc.getAll()
    await oidc.create(request)

    expect(oidc.providers.value.map(p => p.id)).toEqual(['p1', 'p0'])
    expect(oidc.providers.value[0]?.userCount).toBe(0)
  })

  it('toasts a success carrying the provider name', async () => {
    create.mockResolvedValue({ success: true, data: provider })

    await useOidc().create(request)

    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.create.success', color: 'success' }])
  })

  it('toasts a dedicated message on a 409 conflict', async () => {
    create.mockResolvedValue(apiError(409))

    await useOidc().create(request)

    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.create.error', color: 'error' }])
  })

  it('toasts an unknown error on any other failure', async () => {
    create.mockResolvedValue(apiError(500))

    await useOidc().create(request)

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })

  it('leaves the list untouched on failure', async () => {
    create.mockResolvedValue(apiError(500))
    const oidc = useOidc()

    await oidc.create(request)

    expect(oidc.providers.value).toEqual([])
  })
})

describe('useOidc().test', () => {
  it('returns the probe result', async () => {
    testByIssuerUrl.mockResolvedValue({ success: true, data: { issuer: 'https://id.example.org' } })

    await expect(useOidc().test('https://id.example.org')).resolves.toEqual({ issuer: 'https://id.example.org' })
  })

  it('toasts a dedicated message when the issuer is unreachable', async () => {
    testByIssuerUrl.mockResolvedValue(apiError(400, 'issuer is unreachable'))

    await expect(useOidc().test('https://nope.example.org')).resolves.toBeUndefined()
    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.test.error.unreachable', color: 'error' }])
  })

  it('toasts an unknown error on another 400', async () => {
    testByIssuerUrl.mockResolvedValue(apiError(400, 'invalid url'))

    await useOidc().test('nope')

    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })

  it('leaves testLoading false after an unreachable issuer', async () => {
    testByIssuerUrl.mockResolvedValue(apiError(400, 'issuer is unreachable'))
    const oidc = useOidc()

    await oidc.test('https://nope.example.org')

    expect(oidc.testLoading.value).toBe(false)
  })

  it('leaves testLoading false after a success', async () => {
    testByIssuerUrl.mockResolvedValue({ success: true, data: {} })
    const oidc = useOidc()

    await oidc.test('https://id.example.org')

    expect(oidc.testLoading.value).toBe(false)
  })
})

describe('useOidc().deleteById', () => {
  it('drops the provider from the list', async () => {
    getAll.mockResolvedValue({ success: true, data: [provider, { ...provider, id: 'p2' }] })
    deleteById.mockResolvedValue({ success: true, data: undefined })
    const oidc = useOidc()

    await oidc.getAll()
    await oidc.deleteById('p1')

    expect(oidc.providers.value.map(p => p.id)).toEqual(['p2'])
  })

  it('treats a 404 as already deleted', async () => {
    getAll.mockResolvedValue({ success: true, data: [provider] })
    deleteById.mockResolvedValue(apiError(404))
    const oidc = useOidc()

    await oidc.getAll()
    await oidc.deleteById('p1')

    expect(oidc.providers.value).toEqual([])
    expect(useToast().messages.value).toEqual([{ text: 'settings.oidc.delete.success', color: 'success' }])
  })

  it('keeps the provider and toasts an unknown error on a real failure', async () => {
    getAll.mockResolvedValue({ success: true, data: [provider] })
    deleteById.mockResolvedValue(apiError(500))
    const oidc = useOidc()

    await oidc.getAll()
    await oidc.deleteById('p1')

    expect(oidc.providers.value.map(p => p.id)).toEqual(['p1'])
    expect(useToast().messages.value).toEqual([{ text: 'error.unknown', color: 'error' }])
  })
})
