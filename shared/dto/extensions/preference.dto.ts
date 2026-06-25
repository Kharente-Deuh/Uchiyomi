// SPDX-License-Identifier: AGPL-3.0-or-later
type PreferenceTypeDto = 'switch' | 'checkbox' | 'editText' | 'list' | 'multiSelect'

export interface PreferenceDto {
  position: number
  type: PreferenceTypeDto
  key?: string
  title?: string
  summary?: string
  visible: boolean
  booleanValue?: boolean
  booleanDefault?: boolean
  textValue?: string
  textDefault?: string
  multiValue?: string[]
  multiDefault?: string[]
  entries?: string[]
  entryValues?: string[]
  dialogTitle?: string
  dialogMessage?: string
}
