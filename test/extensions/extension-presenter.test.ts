// SPDX-License-Identifier: AGPL-3.0-or-later
import { describe, expect, it } from 'vitest'
import { toExtensionDto, toExtensionListResponseDto, toExtensionSettingsDto, toHealthDto, toPreferenceDto } from '../../server/domains/extensions/infrastructure/transport/http/extension-http.presenter'

describe('extension presenter', () => {
  it('serialises health with ISO dates and log', () => {
    const dto = toHealthDto(
      { pkgName: 'p1', health: 'ERROR', consecutiveFailures: 2, lastErrorAt: new Date('2026-06-22T00:00:00Z'), lastErrorMessage: 'x' },
      [{ occurredAt: new Date('2026-06-22T00:00:00Z'), message: 'x', context: 'install' }],
    )
    expect(dto.lastErrorAt).toBe('2026-06-22T00:00:00.000Z')
    expect(dto.log[0]).toEqual({ occurredAt: '2026-06-22T00:00:00.000Z', message: 'x', context: 'install' })
  })

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

  it('maps isHealthy=true for a healthy installed extension', () => {
    expect(toExtensionDto({ ...base, isHealthy: true })).toMatchObject({ pkgName: 'p', isHealthy: true })
  })

  it('omits isHealthy (undefined) for a non-installed extension', () => {
    expect(toExtensionDto({ ...base, isInstalled: false, isHealthy: undefined }).isHealthy).toBeUndefined()
  })

  it('rewrites iconUrl to the BFF proxy route (never the raw Suwayomi path)', () => {
    const dto = toExtensionDto({ ...base, iconUrl: '/api/v1/extension/icon/tachiyomi-x-v1.0.apk' })
    expect(dto.iconUrl).toBe('/api/extensions/p/icon')
  })

  it('leaves iconUrl undefined when the extension has no icon', () => {
    expect(toExtensionDto({ ...base, iconUrl: undefined }).iconUrl).toBeUndefined()
  })
})

describe('toExtensionListResponseDto', () => {
  it('maps a domain page result into the paginated DTO', () => {
    const dto = toExtensionListResponseDto({
      items: [{ pkgName: 'p', name: 'N', lang: 'en', iconUrl: undefined, isNsfw: false, isInstalled: true, hasUpdate: false, versionName: '1.0', isHealthy: true }],
      page: 2,
      pageSize: 20,
      totalCount: 57,
    })
    expect(dto).toMatchObject({ page: 2, pageSize: 20, totalCount: 57 })
    expect(dto.items.map(i => [i.pkgName, i.isHealthy])).toEqual([['p', true]])
  })
})

describe('toExtensionSettingsDto', () => {
  it('serialises common + per-source preferences', () => {
    const dto = toExtensionSettingsDto({
      pkgName: 'p',
      common: [{ position: 0, type: 'switch', key: 'k', visible: true, booleanValue: true, booleanDefault: false }],
      sources: [{ id: 'a', name: 'A', lang: 'fr', preferences: [{ position: 1, type: 'editText', key: 'token', visible: true, textValue: 'x' }] }],
    })
    expect(dto.pkgName).toBe('p')
    expect(dto.common.map(p => p.key)).toEqual(['k'])
    expect(dto.sources).toEqual([{ id: 'a', name: 'A', lang: 'fr', preferences: [{ position: 1, type: 'editText', key: 'token', visible: true, textValue: 'x' }] }])
  })

  it('drops non-visible preferences and sources left with none', () => {
    const dto = toExtensionSettingsDto({
      pkgName: 'p',
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
