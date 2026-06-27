// SPDX-License-Identifier: AGPL-3.0-or-later
export interface SourceDto {
  id: string
  name: string
  lang: string
  isNsfw: boolean
  isConfigurable: boolean
  isEnabled: boolean
  supportsLatest: boolean
}
