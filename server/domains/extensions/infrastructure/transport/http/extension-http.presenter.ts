// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ExtensionDto, ExtensionHealthDto, ExtensionListResponseDto } from '#shared/dto/extensions/extension.dto'
import type { PreferenceDto } from '#shared/dto/extensions/preference.dto'
import type { SourceDto } from '#shared/dto/extensions/source.dto'
import type { ExtensionErrorLogEntry, ExtensionHealthRow, ExtensionSourcePreferenceModel, ListedExtension, StoredExtensionSource } from '../../../extension.domain'

export function toExtensionDto(e: ListedExtension): ExtensionDto {
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
    isHealthy: e.isHealthy,
  }
}

export function toExtensionListResponseDto(result: {
  items: ListedExtension[]
  page: number
  pageSize: number
  totalCount: number
}): ExtensionListResponseDto {
  return {
    items: result.items.map(e => toExtensionDto(e)),
    page: result.page,
    pageSize: result.pageSize,
    totalCount: result.totalCount,
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
  }
}

export function toPreferenceDto(p: ExtensionSourcePreferenceModel): PreferenceDto {
  return { ...p }
}

export function toHealthDto(h: ExtensionHealthRow, log: ExtensionErrorLogEntry[]): ExtensionHealthDto {
  return {
    pkgName: h.pkgName,
    health: h.health,
    consecutiveFailures: h.consecutiveFailures,
    lastErrorAt: h.lastErrorAt?.toISOString(),
    lastErrorMessage: h.lastErrorMessage,
    log: log.map(e => ({
      occurredAt: e.occurredAt.toISOString(),
      message: e.message,
      context: e.context ?? undefined,
    })),
  }
}
