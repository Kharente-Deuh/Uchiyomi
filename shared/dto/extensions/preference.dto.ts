// SPDX-License-Identifier: AGPL-3.0-or-later

interface BasePreferenceDto {
  position: number
  key?: string
  title?: string
  summary?: string
  visible: boolean
}

interface SwitchPreferenceDto extends BasePreferenceDto {
  type: 'switch'
  booleanValue?: boolean
  booleanDefault: boolean
}

interface CheckboxPreferenceDto extends BasePreferenceDto {
  type: 'checkbox'
  booleanValue?: boolean
  booleanDefault: boolean
}

interface EditTextPreferenceDto extends BasePreferenceDto {
  type: 'editText'
  textValue?: string
  textDefault?: string
  dialogTitle?: string
  dialogMessage?: string
}

interface ListPreferenceDto extends BasePreferenceDto {
  type: 'list'
  textValue?: string
  textDefault?: string
  entries: string[]
  entryValues: string[]
}

interface MultiSelectPreferenceDto extends BasePreferenceDto {
  type: 'multiSelect'
  multiValue?: string[]
  multiDefault?: string[]
  entries: string[]
  entryValues: string[]
}

export type PreferenceDto = SwitchPreferenceDto | CheckboxPreferenceDto | EditTextPreferenceDto | ListPreferenceDto | MultiSelectPreferenceDto
