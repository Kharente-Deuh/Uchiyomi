// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionDto } from '../../shared/dto/extensions/extension.dto'
import { describe, expect, it } from 'vitest'
import { toExtensionDto, toExtensionSettingsDto, toPreferenceDto, toSourceDto } from '../../server/domains/extensions/infrastructure/transport/http/extension-http.presenter'
import { toPageDto } from '../../server/shared'

describe('extension presenter', () => {
  it('passes preference fields through', () => {
    const dto = toPreferenceDto({ position: 0, type: 'switch', visible: true, booleanValue: true, booleanDefault: false })
    expect(dto).toMatchObject({ position: 0, type: 'switch', booleanValue: true, booleanDefault: false })
  })
})

describe('toExtensionDto', () => {
  const base = {
    pkgName: 'p',
    name: 'N',
    lang: 'en',
    iconUrl: undefined,
    isNsfw: false,
    isInstalled: true,
    hasUpdate: false,
    versionName: '1.0',
  }

  it('rewrites iconUrl to the BFF proxy route (never the raw Suwayomi path)', () => {
    const dto = toExtensionDto({ ...base, iconUrl: '/api/v1/extension/icon/tachiyomi-x-v1.0.apk' })
    expect(dto.iconUrl).toBe('/api/extensions/p/icon')
  })

  it('leaves iconUrl undefined when the extension has no icon', () => {
    expect(toExtensionDto({ ...base, iconUrl: undefined }).iconUrl).toBeUndefined()
  })
})

describe('toPageDto', () => {
  it('maps a domain page result into the paginated DTO', () => {
    const dto = toPageDto({
      items: [{ pkgName: 'p', name: 'N', lang: 'en', iconUrl: undefined, isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0' }],
      total: 57,
    }, toExtensionDto)
    expect(dto).toMatchObject({ total: 57 })

    expect(dto.items.map((i: ExtensionDto) => i.pkgName)).toEqual(['p'])
  })
})

describe('toSourceDto', () => {
  const baseSource = {
    id: 'src1',
    pkgName: 'pkg.ext',
    name: 'My Source',
    lang: 'en',
    isNsfw: false,
    isConfigurable: true,
    isEnabled: true,
  }

  it('includes supportsLatest: true', () => {
    const dto = toSourceDto({ ...baseSource, supportsLatest: true })
    expect(dto).toMatchObject({ id: 'src1', name: 'My Source', isEnabled: true, supportsLatest: true })
  })

  it('includes supportsLatest: false', () => {
    const dto = toSourceDto({ ...baseSource, supportsLatest: false })
    expect(dto.supportsLatest).toBe(false)
  })
})

describe('toExtensionSettingsDto', () => {
  it('serialises common + per-source preferences', () => {
    const dto = toExtensionSettingsDto({
      common: [{ position: 0, type: 'switch', key: 'k', visible: true, booleanValue: true, booleanDefault: false }],
      sources: [{ id: 'a', name: 'A', lang: 'fr', preferences: [{ position: 1, type: 'editText', key: 'token', visible: true, textValue: 'x' }] }],
    })
    expect(dto.common.map(p => p.key)).toEqual(['k'])
    expect(dto.sources).toEqual([{ id: 'a', name: 'A', lang: 'fr', preferences: [{ position: 1, type: 'editText', key: 'token', visible: true, textValue: 'x' }] }])
  })

  it('drops non-visible preferences and sources left with none', () => {
    const dto = toExtensionSettingsDto({
      common: [
        { position: 0, type: 'switch', key: 'shown', visible: true, booleanValue: true, booleanDefault: false },
        { position: 1, type: 'switch', key: 'hidden', visible: false, booleanValue: true, booleanDefault: false },
      ],
      sources: [
        { id: 'a', name: 'A', lang: 'fr', preferences: [{ position: 0, type: 'editText', key: 'token', visible: true, textValue: 'x' }] },
        { id: 'b', name: 'B', lang: 'en', preferences: [{ position: 0, type: 'editText', key: 'secret', visible: false, textValue: 'y' }] },
      ],
    })
    expect(dto.common.map(p => p.key)).toEqual(['shown'])
    expect(dto.sources.map(s => s.id)).toEqual(['a'])
  })
})
