// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// --- Mocks hoisted before module imports ---
const { mockApi } = vi.hoisted(() => ({
  mockApi: {
    getExtension: vi.fn(),
    getSettings: vi.fn(),
    listExtensions: vi.fn(),
    extensionAction: vi.fn(),
    listSources: vi.fn(),
    setSourceEnabled: vi.fn(),
    updateSettings: vi.fn(),
  },
}))

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => mockApi,
}))

const { navigateToMock } = vi.hoisted(() => ({
  navigateToMock: vi.fn(),
}))

mockNuxtImport('navigateTo', () => navigateToMock)

const { default: singleExtMiddleware } = await import('~/middleware/single-extension.middleware')
const { useSingleExtensionStore } = await import('~/features/extensions/store/single-extension.store')

// --- fixture data ---
const ext: ExtensionDto = {
  pkgName: 'eu.kanade.test',
  name: 'Test',
  lang: 'en',
  isNsfw: false,
  isInstalled: true,
  hasUpdate: false,
  versionName: '1.0',
}

const settingsWithPrefs: ExtensionSettingsDto = {
  common: [{ type: 'switch', position: 0, visible: true, booleanDefault: false }],
  sources: [],
}

const settingsEmpty: ExtensionSettingsDto = {
  common: [],
  sources: [],
}

function makeRoute(pkgName: string, routeName: string = 'browse-extensions-pkgName'): Parameters<typeof singleExtMiddleware>[0] {
  return {
    path: `/browse/extensions/${pkgName}`,
    fullPath: `/browse/extensions/${pkgName}`,
    params: { pkgName },
    name: routeName,
  } as Parameters<typeof singleExtMiddleware>[0]
}

function makeSettingsRoute(pkgName: string): Parameters<typeof singleExtMiddleware>[0] {
  return makeRoute(pkgName, 'browse-extensions-pkgName-settings')
}

describe('single-extension middleware', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('is a no-op when the store already holds the same extension (non-settings route)', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    await singleExtMiddleware(makeRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(mockApi.getExtension).not.toHaveBeenCalled()
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('is a no-op on settings route when store already has extension and settings', async () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSettings(settingsWithPrefs)
    // For settings route: store.extension.pkgName matches → early return before settings check
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(mockApi.getExtension).not.toHaveBeenCalled()
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('redirects to /browse/extensions when getExtension fails', async () => {
    mockApi.getExtension.mockResolvedValue({ success: false, error: { message: 'not found' } })
    await singleExtMiddleware(makeRoute('unknown.pkg'), makeRoute('unknown.pkg'))
    expect(navigateToMock).toHaveBeenCalledWith('/browse/extensions')
  })

  it('fetches and stores the extension for a non-settings route without calling getSettings', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    const store = useSingleExtensionStore()
    await singleExtMiddleware(makeRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(store.extension).toEqual(ext)
    expect(mockApi.getSettings).not.toHaveBeenCalled()
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('redirects to the extension page when getSettings fails on settings route', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    mockApi.getSettings.mockResolvedValue({ success: false, error: { message: 'fail' } })
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(navigateToMock).toHaveBeenCalledWith(`/browse/extensions/${ext.pkgName}`)
  })

  it('redirects to the extension page when settings are empty (no preferences)', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    mockApi.getSettings.mockResolvedValue({ success: true, data: settingsEmpty })
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(navigateToMock).toHaveBeenCalledWith(`/browse/extensions/${ext.pkgName}`)
  })

  it('stores settings and allows navigation when settings have preferences', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    mockApi.getSettings.mockResolvedValue({ success: true, data: settingsWithPrefs })
    const store = useSingleExtensionStore()
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(store.settings).toEqual(settingsWithPrefs)
    expect(navigateToMock).not.toHaveBeenCalled()
  })

  it('redirects to extension page when settings have only empty source preferences', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    const settingsSourceEmpty: ExtensionSettingsDto = {
      common: [],
      sources: [{ id: 's1', name: 'S', lang: 'en', preferences: [] }],
    }
    mockApi.getSettings.mockResolvedValue({ success: true, data: settingsSourceEmpty })
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(navigateToMock).toHaveBeenCalledWith(`/browse/extensions/${ext.pkgName}`)
  })

  it('allows navigation when settings have source preferences', async () => {
    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    const settingsSourcePrefs: ExtensionSettingsDto = {
      common: [],
      sources: [{ id: 's1', name: 'S', lang: 'en', preferences: [{ type: 'switch', position: 0, visible: true, booleanDefault: true }] }],
    }
    mockApi.getSettings.mockResolvedValue({ success: true, data: settingsSourcePrefs })
    const store = useSingleExtensionStore()
    await singleExtMiddleware(makeSettingsRoute(ext.pkgName), makeRoute(ext.pkgName))
    expect(store.settings).toEqual(settingsSourcePrefs)
    expect(navigateToMock).not.toHaveBeenCalled()
  })
})
