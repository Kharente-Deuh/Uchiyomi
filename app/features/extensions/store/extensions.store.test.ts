// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto } from '#shared/dto/extensions'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useExtensionsStore } from '~/features/extensions/store/extensions.store'

function makeExtension(pkgName: string, overrides: Partial<ExtensionDto> = {}): ExtensionDto {
  return {
    pkgName,
    name: pkgName,
    lang: 'en',
    isNsfw: false,
    isInstalled: false,
    hasUpdate: false,
    versionName: '1.0',
    ...overrides,
  }
}

describe('useExtensionsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  // --- initial state ---

  it('starts with an empty extensions list', () => {
    const store = useExtensionsStore()
    expect(store.extensions).toEqual([])
  })

  it('starts on page 1', () => {
    const store = useExtensionsStore()
    expect(store.page).toBe(1)
  })

  it('starts with all filters undefined', () => {
    const store = useExtensionsStore()
    expect(store.nsfwFilter).toBeUndefined()
    expect(store.isInstalledFilter).toBeUndefined()
    expect(store.isUpToDateFilter).toBeUndefined()
  })

  it('starts with hasFiltersBeenSet false', () => {
    const store = useExtensionsStore()
    expect(store.hasFiltersBeenSet).toBe(false)
  })

  // --- setExtensions ---

  it('setExtensions replaces the extensions list', () => {
    const store = useExtensionsStore()
    const exts = [makeExtension('pkg.a'), makeExtension('pkg.b')]
    store.setExtensions(exts)
    expect(store.extensions).toHaveLength(2)
    expect(store.extensions[0]!.pkgName).toBe('pkg.a')
    expect(store.extensions[1]!.pkgName).toBe('pkg.b')
  })

  // --- update ---

  it('update replaces an existing extension by pkgName', () => {
    const store = useExtensionsStore()
    store.setExtensions([makeExtension('pkg.a', { isInstalled: false })])
    store.update(makeExtension('pkg.a', { isInstalled: true }))
    expect(store.extensions[0]!.isInstalled).toBe(true)
  })

  it('update is a no-op when pkgName does not exist', () => {
    const store = useExtensionsStore()
    store.setExtensions([makeExtension('pkg.a')])
    store.update(makeExtension('pkg.unknown'))
    expect(store.extensions).toHaveLength(1)
    expect(store.extensions[0]!.pkgName).toBe('pkg.a')
  })

  // --- setPage ---

  it('setPage updates the page value', () => {
    const store = useExtensionsStore()
    store.setPage(3)
    expect(store.page).toBe(3)
  })

  // --- filter setters: set value AND flip hasFiltersBeenSet ---

  it('setNsfwFilter sets nsfwFilter and marks hasFiltersBeenSet', () => {
    const store = useExtensionsStore()
    store.setNsfwFilter(true)
    expect(store.nsfwFilter).toBe(true)
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  it('setNsfwFilter can be cleared to undefined', () => {
    const store = useExtensionsStore()
    store.setNsfwFilter(true)
    store.setNsfwFilter(undefined)
    expect(store.nsfwFilter).toBeUndefined()
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  it('setIsInstalledFilter sets isInstalledFilter and marks hasFiltersBeenSet', () => {
    const store = useExtensionsStore()
    store.setIsInstalledFilter(true)
    expect(store.isInstalledFilter).toBe(true)
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  it('setIsInstalledFilter can be cleared to undefined', () => {
    const store = useExtensionsStore()
    store.setIsInstalledFilter(true)
    store.setIsInstalledFilter(undefined)
    expect(store.isInstalledFilter).toBeUndefined()
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  it('setIsUpToDateFilter sets isUpToDateFilter and marks hasFiltersBeenSet', () => {
    const store = useExtensionsStore()
    store.setIsUpToDateFilter(false)
    expect(store.isUpToDateFilter).toBe(false)
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  it('setIsUpToDateFilter can be cleared to undefined', () => {
    const store = useExtensionsStore()
    store.setIsUpToDateFilter(false)
    store.setIsUpToDateFilter(undefined)
    expect(store.isUpToDateFilter).toBeUndefined()
    expect(store.hasFiltersBeenSet).toBe(true)
  })

  // --- clear ---

  it('clear empties the extensions list', () => {
    const store = useExtensionsStore()
    store.setExtensions([makeExtension('pkg.a'), makeExtension('pkg.b')])
    store.clear()
    expect(store.extensions).toEqual([])
  })
})
