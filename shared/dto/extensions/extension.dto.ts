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
}
