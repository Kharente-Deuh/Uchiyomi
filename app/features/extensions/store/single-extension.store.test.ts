// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionDto, SourceDto } from '#shared/dto/extensions'
import type { ExtensionSettingsDto } from '#shared/dto/extensions/extension-settings.dto'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useSingleExtensionStore } from '~/features/extensions/store/single-extension.store'

const ext: ExtensionDto = {
  pkgName: 'eu.kanade.tachiyomi.extension.en.mangadex',
  name: 'MangaDex',
  lang: 'en',
  isNsfw: false,
  isInstalled: true,
  hasUpdate: false,
  versionName: '1.5.0',
}

const source1: SourceDto = {
  id: 's1',
  name: 'MangaDex',
  lang: 'en',
  isNsfw: false,
  isConfigurable: false,
  isEnabled: true,
  supportsLatest: true,
}

const source2: SourceDto = {
  id: 's2',
  name: 'MangaDex Low Quality',
  lang: 'en',
  isNsfw: false,
  isConfigurable: false,
  isEnabled: false,
  supportsLatest: false,
}

const settings: ExtensionSettingsDto = {
  common: [],
  sources: [],
}

describe('useSingleExtensionStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('starts with no extension, empty sources, and no settings', () => {
    const store = useSingleExtensionStore()
    expect(store.extension).toBeUndefined()
    expect(store.sources).toEqual([])
    expect(store.settings).toBeUndefined()
  })

  it('setExtension stores the extension', () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    expect(store.extension).toEqual(ext)
  })

  it('setSources replaces the sources array', () => {
    const store = useSingleExtensionStore()
    store.setSources([source1, source2])
    expect(store.sources).toHaveLength(2)
    expect(store.sources[0]!.id).toBe('s1')
    expect(store.sources[1]!.id).toBe('s2')
  })

  it('setSources with empty array clears sources', () => {
    const store = useSingleExtensionStore()
    store.setSources([source1])
    store.setSources([])
    expect(store.sources).toHaveLength(0)
  })

  it('updateSource replaces the matching source by id', () => {
    const store = useSingleExtensionStore()
    store.setSources([source1, source2])
    const updated: SourceDto = { ...source1, isEnabled: false }
    store.updateSource(updated)
    expect(store.sources[0]!.isEnabled).toBe(false)
    // The other source is unchanged
    expect(store.sources[1]!.isEnabled).toBe(false)
  })

  it('updateSource is a no-op when the source id is not found', () => {
    const store = useSingleExtensionStore()
    store.setSources([source1])
    const unknown: SourceDto = { ...source2, id: 'unknown' }
    store.updateSource(unknown)
    expect(store.sources).toHaveLength(1)
    expect(store.sources[0]!.id).toBe('s1')
  })

  it('setSettings stores the settings', () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings)
    expect(store.settings).toEqual(settings)
  })

  it('updateSettings replaces the settings', () => {
    const store = useSingleExtensionStore()
    store.setSettings(settings)
    const newSettings: ExtensionSettingsDto = {
      common: [{ type: 'switch', position: 0, key: 'pref1', visible: true, booleanDefault: false }],
      sources: [],
    }
    store.updateSettings(newSettings)
    expect(store.settings!.common).toHaveLength(1)
  })

  it('clear resets extension, sources, and settings to initial state', () => {
    const store = useSingleExtensionStore()
    store.setExtension(ext)
    store.setSources([source1])
    store.setSettings(settings)
    store.clear()
    expect(store.extension).toBeUndefined()
    expect(store.sources).toEqual([])
    expect(store.settings).toBeUndefined()
  })
})
