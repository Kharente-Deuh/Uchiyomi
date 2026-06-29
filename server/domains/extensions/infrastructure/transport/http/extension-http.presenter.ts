// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ExtensionSettingsDto, ExtensionSettingsSourceDto } from '#shared/dto/extensions/extension-settings.dto'
import type { ExtensionDto } from '#shared/dto/extensions/extension.dto'
import type { PreferenceDto } from '#shared/dto/extensions/preference.dto'
import type { SourceDto } from '#shared/dto/extensions/source.dto'
import type { ExtensionModel, ExtensionSettings, ExtensionSourcePreferenceModel, StoredExtensionSource } from '../../../extension.domain'

export function toExtensionDto(e: ExtensionModel): ExtensionDto {
  return {
    pkgName: e.pkgName,
    name: e.name,
    lang: e.lang,
    // Suwayomi serves the icon from an internal `/api/v1/extension/icon/<apk>`
    // route (image bytes despite the `.apk` path). Suwayomi is never exposed
    // (ADR-0001), so we hand the client a BFF proxy URL rather than the raw path.
    iconUrl: e.iconUrl ? `/api/extensions/${e.pkgName}/icon` : undefined,
    isNsfw: e.isNsfw,
    isInstalled: e.isInstalled,
    hasUpdate: e.hasUpdate,
    versionName: e.versionName,
  }
}

export function toSourceDto(s: StoredExtensionSource): SourceDto {
  return {
    id: s.id,
    name: s.name,
    lang: s.lang,
    isNsfw: s.isNsfw,
    isConfigurable: s.isConfigurable,
    isEnabled: s.isEnabled,
    supportsLatest: s.supportsLatest,
  }
}

export function toPreferenceDto(p: ExtensionSourcePreferenceModel): PreferenceDto {
  // The domain and DTO branches are structurally identical; spreading the
  // narrowed branch satisfies the matching DTO branch of the union.
  switch (p.type) {
    case 'switch':
      return { ...p }
    case 'checkbox':
      return { ...p }
    case 'editText':
      return { ...p }
    case 'list':
      return { ...p }
    case 'multiSelect':
      return { ...p }
  }
}

export function toExtensionSettingsDto(s: ExtensionSettings): ExtensionSettingsDto {
  const commonPreferences: PreferenceDto[] = []
  for (const preference of s.common) {
    if (preference.visible) {
      commonPreferences.push(toPreferenceDto(preference))
    }
  }

  const mappedSources: ExtensionSettingsSourceDto[] = []
  for (const source of s.sources) {
    const sourcePreferences: PreferenceDto[] = []
    for (const preference of source.preferences) {
      if (preference.visible) {
        sourcePreferences.push(toPreferenceDto(preference))
      }
    }

    if (sourcePreferences.length > 0) {
      mappedSources.push({
        id: source.id,
        name: source.name,
        lang: source.lang,
        preferences: sourcePreferences,
      })
    }
  }

  return {
    common: commonPreferences,
    sources: mappedSources,
  }
}
