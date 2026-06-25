// SPDX-License-Identifier: AGPL-3.0-or-later
import { afterEach, expect, it, vi } from 'vitest'

const apiFetch = vi.fn()
vi.mock('~/utils/api/api-fetch', () => ({ apiFetch }))

const { createExtensionsApi } = await import('~/features/extensions/api/extensions.api')
const { ApiError } = await import('~/utils/api')

afterEach(() => apiFetch.mockReset())

// --- listExtensions ---

it('listExtensions hits /api/extensions with a params object', async () => {
  const payload = { items: [{ pkgName: 'p', name: 'P' }], totalCount: 1 }
  apiFetch.mockResolvedValueOnce(payload)
  const res = await createExtensionsApi().listExtensions({ page: 1, pageSize: 20 })
  expect(res).toEqual({ success: true, data: payload })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions', { query: { page: 1, pageSize: 20 } })
})

it('listExtensions passes filter params via query option', async () => {
  apiFetch.mockResolvedValueOnce({ items: [], totalCount: 0 })
  await createExtensionsApi().listExtensions({ isInstalled: true, hasUpdate: false })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions', { query: { isInstalled: true, hasUpdate: false } })
})

it('listExtensions passes undefined filter values through (caller filters them)', async () => {
  apiFetch.mockResolvedValueOnce({ items: [], totalCount: 0 })
  await createExtensionsApi().listExtensions({ isInstalled: true, nsfw: undefined })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions', { query: { isInstalled: true, nsfw: undefined } })
})

it('listExtensions maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(Object.assign(new Error('Server Error'), {
    statusCode: 500,
    data: { statusMessage: 'Server Error' },
  }))
  const res = await createExtensionsApi().listExtensions({})
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
    expect(res.error.status).toBe(500)
  }
})

// --- getExtension ---

it('getExtension hits /api/extensions/:pkgName', async () => {
  const payload = { extension: { pkgName: 'p' }, health: null }
  apiFetch.mockResolvedValueOnce(payload)
  const res = await createExtensionsApi().getExtension('p')
  expect(res).toEqual({ success: true, data: payload })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p')
})

it('getExtension maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().getExtension('p')
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- extensionAction ---

it('extensionAction posts to /api/extensions/:pkgName', async () => {
  apiFetch.mockResolvedValueOnce({ ok: true })
  await createExtensionsApi().extensionAction('p', 'install')
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p', { method: 'POST', body: { action: 'install' } })
})

it('extensionAction returns the updated extension on success', async () => {
  const extension = { pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }
  apiFetch.mockResolvedValueOnce(extension)
  const res = await createExtensionsApi().extensionAction('pkg.name', 'install')
  expect(res).toEqual({ success: true, data: extension })
})

it('extensionAction maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().extensionAction('p', 'install')
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- listSources ---

it('listSources hits /api/extensions/:pkgName/sources', async () => {
  apiFetch.mockResolvedValueOnce({ sources: [{ id: 's' }] })
  const res = await createExtensionsApi().listSources('p')
  expect(res).toEqual({ success: true, data: [{ id: 's' }] })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources')
})

it('listSources maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().listSources('p')
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- setSourceEnabled ---

it('setSourceEnabled patches /api/sources/:id', async () => {
  apiFetch.mockResolvedValueOnce({ source: { id: 's', isEnabled: false } })
  await createExtensionsApi().setSourceEnabled('s', false)
  expect(apiFetch).toHaveBeenCalledWith('/api/sources/s', { method: 'PATCH', body: { isEnabled: false } })
})

it('setSourceEnabled returns the updated source', async () => {
  const source = { id: 's', isEnabled: true }
  apiFetch.mockResolvedValueOnce({ source })
  const res = await createExtensionsApi().setSourceEnabled('s', true)
  expect(res).toEqual({ success: true, data: source })
})

it('setSourceEnabled maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().setSourceEnabled('s', true)
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- getPreferences ---

it('getPreferences hits /api/sources/:id/preferences', async () => {
  apiFetch.mockResolvedValueOnce({ preferences: [] })
  await createExtensionsApi().getPreferences('s')
  expect(apiFetch).toHaveBeenCalledWith('/api/sources/s/preferences')
})

it('getPreferences returns data', async () => {
  apiFetch.mockResolvedValueOnce({ preferences: [{ key: 'k' }] })
  const res = await createExtensionsApi().getPreferences('s')
  expect(res).toEqual({ success: true, data: [{ key: 'k' }] })
})

it('getPreferences maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().getPreferences('s')
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- updatePreference ---

it('updatePreference puts to /api/sources/:id/preferences', async () => {
  apiFetch.mockResolvedValueOnce({ preferences: [] })
  await createExtensionsApi().updatePreference('s', { position: 0, booleanValue: true })
  expect(apiFetch).toHaveBeenCalledWith('/api/sources/s/preferences', { method: 'PUT', body: { position: 0, booleanValue: true } })
})

it('updatePreference maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().updatePreference('s', { position: 0 })
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})
