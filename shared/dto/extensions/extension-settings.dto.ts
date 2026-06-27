// SPDX-License-Identifier: AGPL-3.0-or-later
import type { PreferenceDto } from './preference.dto'

export interface ExtensionSettingsSourceDto {
  id: string
  name: string
  lang: string
  preferences: PreferenceDto[]
}

export interface ExtensionSettingsDto {
  pkgName: string
  common: PreferenceDto[]
  sources: ExtensionSettingsSourceDto[]
}
