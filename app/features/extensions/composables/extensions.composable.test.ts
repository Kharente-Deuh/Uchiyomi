import type { Ref } from 'vue'
// SPDX-License-Identifier: AGPL-3.0-or-later
import { mockNuxtImport } from '@nuxt/test-utils/runtime'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const { mockApi } = vi.hoisted(() => ({
  mockApi: {
    listExtensions: vi.fn(),
    getExtension: vi.fn(),
    extensionAction: vi.fn(),
    listSources: vi.fn(),
    setSourceEnabled: vi.fn(),
    getPreferences: vi.fn(),
    updatePreference: vi.fn(),
  },
}))

vi.mock('~/features/extensions/api/extensions.api', () => ({
  createExtensionsApi: () => mockApi,
}))

// Minimal ref-like objects compatible with Vue's Ref contract.
// vi.hoisted runs before module imports so we cannot use vue's ref() here —
// plain objects with a mutable `.value` satisfy the composable's access pattern.
// Typed explicitly as mutable so test assignments to `.value` compile.
const { queryStub } = vi.hoisted(() => ({
  queryStub: {
    data: { value: undefined as unknown } as { value: unknown },
    isLoading: { value: false } as { value: boolean },
  },
}))

vi.mock('@pinia/colada', () => ({
  useQuery: () => queryStub,
}))

// Mock @vueuse/core's useDebounce (used internally for the search filter).
vi.mock('@vueuse/core', async (importOriginal) => {
  const orig = await importOriginal<typeof import('@vueuse/core')>()

  return {
    ...orig,
    useDebounce: (v: { value: unknown }) => v,
  }
})

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

const { useExtensions } = await import('~/features/extensions/composables/extensions.composable')
const { useExtensionsStore } = await import('~/features/extensions/store/extensions.store')
const { ApiError } = await import('~/utils/api')

describe('useExtensions', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    queryStub.data.value = undefined
    queryStub.isLoading.value = false
  })

  // --- reactive filters ---

  it('exposes nsfwFilter as a writable ref defaulting to undefined', () => {
    const { nsfwFilter } = useExtensions()
    expect(nsfwFilter.value).toBeUndefined()
    // Cast to Ref to allow assignment — the interface allows ComputedRef but
    // the implementation always returns a plain Ref, which is writable.
    ;(nsfwFilter as Ref<boolean | undefined>).value = true
    expect(nsfwFilter.value).toBe(true)
  })

  it('exposes isInstalledFilter as a writable ref defaulting to undefined', () => {
    const { isInstalledFilter } = useExtensions()
    expect(isInstalledFilter.value).toBeUndefined()
    ;(isInstalledFilter as Ref<boolean | undefined>).value = true
    expect(isInstalledFilter.value).toBe(true)
  })

  it('exposes isUpToDateFilter as a writable ref defaulting to undefined', () => {
    const { isUpToDateFilter } = useExtensions()
    expect(isUpToDateFilter.value).toBeUndefined()
    ;(isUpToDateFilter as Ref<boolean | undefined>).value = false
    expect(isUpToDateFilter.value).toBe(false)
  })

  it('exposes searchFilter as a writable ref defaulting to undefined', () => {
    const { searchFilter } = useExtensions()
    expect(searchFilter.value).toBeUndefined()
    ;(searchFilter as Ref<string | undefined>).value = 'manga'
    expect(searchFilter.value).toBe('manga')
  })

  // --- extensions from store ---

  it('extensions reflects the store contents', () => {
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'p', name: 'P', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }])
    const { extensions } = useExtensions()
    expect(extensions.value).toHaveLength(1)
    expect(extensions.value[0]!.pkgName).toBe('p')
  })

  // --- fetchLoading ---

  it('fetchLoading starts as false', () => {
    const { fetchLoading } = useExtensions()
    expect(fetchLoading.value).toBe(false)
  })

  // --- maxPage ---

  it('maxPage is 0 when there is no data', () => {
    const { maxPage } = useExtensions()
    expect(maxPage.value).toBe(0)
  })

  // --- install ---

  it('install calls extensionAction with install', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: false, versionName: '1.0' }])
    const { install } = useExtensions()
    await install('pkg.name')
    expect(mockApi.extensionAction).toHaveBeenCalledWith('pkg.name', 'install')
  })

  it('install marks the extension as installed in the store on success', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: false, versionName: '1.0' }])
    const { install } = useExtensions()
    await install('pkg.name')
    expect(store.extensions[0]!.isInstalled).toBe(true)
  })

  it('install removes pkgName from installExtensionsLoading after completion', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const { install, installExtensionsLoading } = useExtensions()
    await install('pkg.name')
    expect(installExtensionsLoading.value).not.toContain('pkg.name')
  })

  it('install shows error toast on failure', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { install } = useExtensions()
    await install('pkg.name')
    expect(errorMock).toHaveBeenCalledWith('extensions.errors.actionFailed')
  })

  it('install does not update the store on failure', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: false, hasUpdate: false, versionName: '1.0' }])
    const { install } = useExtensions()
    await install('pkg.name')
    expect(store.extensions[0]!.isInstalled).toBe(false)
  })

  // --- uninstall ---

  it('uninstall calls extensionAction with uninstall', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }])
    const { uninstall } = useExtensions()
    await uninstall('pkg.name')
    expect(mockApi.extensionAction).toHaveBeenCalledWith('pkg.name', 'uninstall')
  })

  it('uninstall marks the extension as not installed in the store on success', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }])
    const { uninstall } = useExtensions()
    await uninstall('pkg.name')
    expect(store.extensions[0]!.isInstalled).toBe(false)
  })

  it('uninstall resets uninstallLoading to false after completion', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: true, data: undefined })
    const { uninstall, uninstallLoading } = useExtensions()
    await uninstall('pkg.name')
    expect(uninstallLoading.value).toBe(false)
  })

  it('uninstall shows error toast on failure', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const { uninstall } = useExtensions()
    await uninstall('pkg.name')
    expect(errorMock).toHaveBeenCalledWith('extensions.errors.actionFailed')
  })

  it('uninstall does not update the store on failure', async () => {
    mockApi.extensionAction.mockResolvedValue({ success: false, error: new ApiError('fail', 500) })
    const store = useExtensionsStore()
    store.setExtensions([{ pkgName: 'pkg.name', name: 'P', lang: 'en', isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }])
    const { uninstall } = useExtensions()
    await uninstall('pkg.name')
    expect(store.extensions[0]!.isInstalled).toBe(true)
  })
})
