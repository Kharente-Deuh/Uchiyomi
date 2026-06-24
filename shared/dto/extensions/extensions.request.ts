// SPDX-License-Identifier: AGPL-3.0-or-later
export interface ExtensionActionRequestDto {
  action: 'install' | 'uninstall'
}

export interface UpdateSourcePreferenceRequestDto {
  position: number
  booleanValue?: boolean
  textValue?: string
  multiValue?: string[]
}

export interface UpdateSourceRequestDto {
  isEnabled: boolean
}

export interface ExtensionListQueryDto {
  page?: number
  pageSize?: number
  search?: string
  isInstalled?: boolean
  hasUpdate?: boolean
  nsfw?: boolean
}
