// SPDX-License-Identifier: AGPL-3.0-or-later
export interface ExtensionDto {
  pkgName: string
  name: string
  lang: string
  iconUrl?: string
  isNsfw: boolean
  isInstalled: boolean
  hasUpdate: boolean
  versionName: string
  isHealthy?: boolean
}

export interface ExtensionHealthDto {
  pkgName: string
  health: 'OK' | 'ERROR'
  consecutiveFailures: number
  lastErrorAt?: string
  lastErrorMessage?: string
  log: { occurredAt: string, message: string, context?: string }[]
}

export interface ExtensionListResponseDto {
  items: ExtensionDto[]
  page: number
  pageSize: number
  totalCount: number
}
