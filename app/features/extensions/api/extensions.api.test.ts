// SPDX-License-Identifier: AGPL-3.0-or-later

import { afterEach, expect, it, vi } from 'vitest'

const apiFetch = vi.fn()
vi.mock('~/utils/api/api-fetch', () => ({ apiFetch }))

const { createExtensionsApi } = await import('~/features/extensions/api/extensions.api')
const { ApiError } = await import('~/utils/api')

afterEach(() => apiFetch.mockReset())

// --- listExtensions ---

it('listExtensions hits /api/extensions with a params object', async () => {
  const payload = { items: [{ pkgName: 'p', name: 'P' }], total: 1 }
  apiFetch.mockResolvedValueOnce(payload)
  const res = await createExtensionsApi().listExtensions({ page: 1, pageSize: 20 })
  expect(res).toEqual({ success: true, data: payload })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions', { query: { page: 1, pageSize: 20 } })
})

it('listExtensions passes filter params via query option', async () => {
  apiFetch.mockResolvedValueOnce({ items: [], total: 0 })
  await createExtensionsApi().listExtensions({ isInstalled: true, hasUpdate: false })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions', { query: { isInstalled: true, hasUpdate: false } })
})

it('listExtensions passes undefined filter values through (caller filters them)', async () => {
  apiFetch.mockResolvedValueOnce({ items: [], total: 0 })
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
  const payload = { extension: { pkgName: 'p' } }
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

it('setSourceEnabled posts to /api/extensions/:pkgName/sources/:id/enable', async () => {
  apiFetch.mockResolvedValueOnce({ source: { id: 's', isEnabled: false } })
  await createExtensionsApi().setSourceEnabled('p', 's', false)
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources/s/enable', { method: 'POST', body: { isEnabled: false } })
})

it('setSourceEnabled returns the updated source', async () => {
  const source = { id: 's', isEnabled: true }
  apiFetch.mockResolvedValueOnce({ source })
  const res = await createExtensionsApi().setSourceEnabled('p', 's', true)
  expect(res).toEqual({ success: true, data: source })
})

it('setSourceEnabled maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().setSourceEnabled('p', 's', true)
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- getSettings ---

it('getSettings hits /api/extensions/:pkgName/sources/settings', async () => {
  apiFetch.mockResolvedValueOnce({ pkgName: 'p', common: [], sources: [] })
  await createExtensionsApi().getSettings('p')
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources/settings')
})

it('getSettings returns the settings payload', async () => {
  const settings = { pkgName: 'p', common: [{ key: 'k' }], sources: [] }
  apiFetch.mockResolvedValueOnce(settings)
  const res = await createExtensionsApi().getSettings('p')
  expect(res).toEqual({ success: true, data: settings })
})

it('getSettings maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().getSettings('p')
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- updateSettings ---

it('updateSettings puts to /api/extensions/:pkgName/sources/settings', async () => {
  const body = { common: [], sources: [] }
  apiFetch.mockResolvedValueOnce({ pkgName: 'p', common: [], sources: [] })
  await createExtensionsApi().updateSettings('p', body)
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources/settings', { method: 'PUT', body })
})

it('updateSettings returns the updated settings', async () => {
  const settings = { pkgName: 'p', common: [], sources: [] }
  apiFetch.mockResolvedValueOnce(settings)
  const res = await createExtensionsApi().updateSettings('p', { common: [], sources: [] })
  expect(res).toEqual({ success: true, data: settings })
})

it('updateSettings maps fetch errors to ApiError', async () => {
  apiFetch.mockRejectedValueOnce(new Error('nope'))
  const res = await createExtensionsApi().updateSettings('p', { common: [], sources: [] })
  expect(res.success).toBe(false)
  if (!res.success) {
    expect(res.error).toBeInstanceOf(ApiError)
  }
})

// --- getMangaChapterSummary ---

it('getMangaChapterSummary hits the mangas chapter-summary route', async () => {
  const payload = { chapterCount: 3, lastChapter: null }
  apiFetch.mockResolvedValueOnce(payload)
  const res = await createExtensionsApi().getMangaChapterSummary('p', '9', '42')
  expect(res).toEqual({ success: true, data: payload })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources/9/mangas/42/chapter-summary', { signal: undefined })
})

it('getMangaChapterSummary forwards an abort signal', async () => {
  apiFetch.mockResolvedValueOnce({ chapterCount: 0, lastChapter: null })
  const controller = new AbortController()
  await createExtensionsApi().getMangaChapterSummary('p', '9', '42', { signal: controller.signal })
  expect(apiFetch).toHaveBeenCalledWith('/api/extensions/p/sources/9/mangas/42/chapter-summary', { signal: controller.signal })
})
