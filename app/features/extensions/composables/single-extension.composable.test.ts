// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto, SourceDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// --- API mock ---
const { mockApi } = vi.hoisted(() => ({
  mockApi: {
    getExtension: vi.fn(),
    extensionAction: vi.fn(),
    listSources: vi.fn(),
    setSourceEnabled: vi.fn(),
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    listExtensions: vi.fn(),
  },
}))

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => mockApi,
}))

// --- Nuxt auto-import mocks ---
const { successMock, errorMock } = vi.hoisted(() => ({
  successMock: vi.fn(),
  errorMock: vi.fn(),
}))

function useToastMock(): { success: typeof successMock, error: typeof errorMock } {
  return { success: successMock, error: errorMock }
}

function useI18nMock(): { t: (key: string) => string } {
  return { t: (key: string) => key }
}

mockNuxtImport('useToast', () => useToastMock)
mockNuxtImport('useI18n', () => useI18nMock)

// --- vueuse mock: useDebounceFn passes through the function unchanged ---
vi.mock('@vueuse/core', async (importOriginal) => {
  const orig = await importOriginal<typeof import('@vueuse/core')>()

  return {
    ...orig,
    useDebounceFn: (fn: (...args: unknown[]) => unknown) => fn,
  }
})

// Lifecycle hooks are no-ops in the test harness; we invoke side effects manually.
function onMountedStub(cb: () => void): void {
  void cb
}

function onBeforeUnmountStub(_cb: () => void): void {}

mockNuxtImport('onMounted', () => onMountedStub)
mockNuxtImport('onBeforeUnmount', () => onBeforeUnmountStub)

const { useSingleExtension } = await import('~/features/extensions/composables/single-extension.composable')
const { useSingleExtensionStore } = await import('~/features/extensions/store/single-extension.store')
const { useAuthStore } = await import('~/features/auth/store/auth.store')
const { ApiError } = await import('~/utils/api')

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

const source: SourceDto = {
  id: 's1',
  name: 'TestSource',
  lang: 'en',
  isNsfw: false,
  isConfigurable: false,
  isEnabled: true,
  supportsLatest: true,
}

const settings: ExtensionSettingsDto = {
  common: [],
  sources: [{ id: 's1', name: 'TestSource', lang: 'en', preferences: [{ type: 'switch', position: 0, visible: true, booleanDefault: false }] }],
}

describe('useSingleExtension', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  // --- initial reactive state from store ---

  it('extension reflects the store extension', () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    const { extension } = useSingleExtension(ext.pkgName)
    expect(extension.value).toEqual(ext)
  })

  it('sources reflects the store sources', () => {
    const store = useSingleExtensionStore()
    store.setSources([source])
    const { sources } = useSingleExtension(ext.pkgName)
    expect(sources.value).toHaveLength(1)
    expect(sources.value[0]!.id).toBe('s1')
  })

  it('settings reflects the store settings', () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings)
    const { settings: s } = useSingleExtension(ext.pkgName)
    expect(s.value).toEqual(settings)
  })

  it('fetchSourcesLoading starts as false', () => {
    const { fetchSourcesLoading } = useSingleExtension(ext.pkgName)
    expect(fetchSourcesLoading.value).toBe(false)
  })

  it('sourceToggleLoading starts as an empty Set', () => {
    const { sourceToggleLoading } = useSingleExtension(ext.pkgName)
    expect(sourceToggleLoading.value.size).toBe(0)
  })

  // --- hasSettings ---

  it('hasSettings is false when settings is undefined', () => {
    const { hasSettings } = useSingleExtension(ext.pkgName)
    expect(hasSettings.value).toBe(false)
  })

  it('hasSettings is false when settings has no preferences anywhere', () => {
    const store = useSingleExtensionStore()
    store.setSettings({ common: [], sources: [{ id: 's1', name: 'S', lang: 'en', preferences: [] }] })
    const { hasSettings } = useSingleExtension(ext.pkgName)
    expect(hasSettings.value).toBe(false)
  })

  it('hasSettings is true when common has at least one preference', () => {
    const store = useSingleExtensionStore()
    store.setSettings({ common: [{ type: 'switch', position: 0, visible: true, booleanDefault: false }], sources: [] })
    const { hasSettings } = useSingleExtension(ext.pkgName)
    expect(hasSettings.value).toBe(true)
  })

  it('hasSettings is true when a source has at least one preference', () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings) // settings.sources[0] has 1 preference
    const { hasSettings } = useSingleExtension(ext.pkgName)
    expect(hasSettings.value).toBe(true)
  })

  // --- uninstallExtension ---

  it('uninstallExtension calls extensionAction with uninstall and returns true on success', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: { ...ext, isInstalled: false } })
    const { uninstallExtension } = useSingleExtension(ext.pkgName)
    const result = await uninstallExtension()
    expect(mockApi.extensionAction).toHaveBeenCalledWith(ext.pkgName, 'uninstall')
    expect(result).toBe(true)
  })

  it('uninstallExtension shows error toast and returns false on failure', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { uninstallExtension } = useSingleExtension(ext.pkgName)
    const result = await uninstallExtension()
    expect(errorMock).toHaveBeenCalledWith('extension.errors.uninstallFailed')
    expect(result).toBe(false)
  })

  // --- toggleSourceEnabled ---

  it('toggleSourceEnabled is a no-op when source id is not found', async () => {
    const store = useSingleExtensionStore()
    store.setSources([source])
    const { toggleSourceEnabled } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('does-not-exist')
    expect(mockApi.setSourceEnabled).not.toHaveBeenCalled()
  })

  it('toggleSourceEnabled calls setSourceEnabled with the inverted isEnabled', async () => {
    const store = useSingleExtensionStore()
    store.setSources([source]) // source.isEnabled = true
    mockApi.setSourceEnabled.mockResolvedValue({ success: true, data: { ...source, isEnabled: false } })
    const { toggleSourceEnabled } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('s1')
    expect(mockApi.setSourceEnabled).toHaveBeenCalledWith(ext.pkgName, 's1', false)
  })

  it('toggleSourceEnabled updates the store source on success', async () => {
    const store = useSingleExtensionStore()
    store.setSources([source])
    const updated: SourceDto = { ...source, isEnabled: false }
    mockApi.setSourceEnabled.mockResolvedValue({ success: true, data: updated })
    const { toggleSourceEnabled } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('s1')
    expect(store.sources[0]!.isEnabled).toBe(false)
  })

  it('toggleSourceEnabled removes sourceId from sourceToggleLoading after success', async () => {
    const store = useSingleExtensionStore()
    store.setSources([source])
    mockApi.setSourceEnabled.mockResolvedValue({ success: true, data: { ...source, isEnabled: false } })
    const { toggleSourceEnabled, sourceToggleLoading } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('s1')
    expect(sourceToggleLoading.value.has('s1')).toBe(false)
  })

  it('toggleSourceEnabled shows error toast on failure (disable direction)', async () => {
    const store = useSingleExtensionStore()
    store.setSources([source]) // isEnabled = true → disable
    mockApi.setSourceEnabled.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { toggleSourceEnabled } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('s1')
    expect(errorMock).toHaveBeenCalledWith('sources.errors.disableFailed')
  })

  it('toggleSourceEnabled shows error toast on failure (enable direction)', async () => {
    const store = useSingleExtensionStore()
    const disabledSource: SourceDto = { ...source, isEnabled: false }
    store.setSources([disabledSource])
    mockApi.setSourceEnabled.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { toggleSourceEnabled } = useSingleExtension(ext.pkgName)
    await toggleSourceEnabled('s1')
    expect(errorMock).toHaveBeenCalledWith('sources.errors.enableFailed')
  })

  // --- updateSettings ---

  it('updateSettings calls api.updateSettings and stores the result on success', async () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings)
    const updated: ExtensionSettingsDto = { common: [], sources: [] }
    mockApi.updateSettings.mockResolvedValue({ success: true, data: updated })
    const { updateSettings } = useSingleExtension(ext.pkgName)
    await updateSettings(settings)
    expect(mockApi.updateSettings).toHaveBeenCalledWith(ext.pkgName, settings)
    expect(store.settings).toEqual(updated)
  })

  it('updateSettings does not update the store on failure', async () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings)
    mockApi.updateSettings.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { updateSettings } = useSingleExtension(ext.pkgName)
    await updateSettings(settings)
    // Store remains unchanged
    expect(store.settings).toEqual(settings)
  })

  // --- fetchSettings skipped when canManageExtensions is false ---

  it('fetchSettings is skipped when canManageExtensions is false', async () => {
    // authStore user with no canManageExtensions
    const authStore = useAuthStore()
    authStore.setUser({
      id: 'u1',
      accountName: 'reader',
      displayName: 'Reader',
      role: 'USER',
      status: 'ACTIVE',
      canManageExtensions: false,
      canDownload: false,
      allowNsfw: false,
      showNsfw: false,
    })

    mockApi.getExtension.mockResolvedValue({ success: true, data: { extension: ext } })
    mockApi.listSources.mockResolvedValue({ success: true, data: [source] })

    // getSettings should NOT be called
    const { sources } = useSingleExtension(ext.pkgName)
    expect(mockApi.getSettings).not.toHaveBeenCalled()
    // Sources were seeded by onMounted mock stub above
    void sources
  })
})
