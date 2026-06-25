// SPDX-License-Identifier: AGPL-3.0-or-later
import type { GetSourcePreferencesQuery } from '../../../../../utils/suwayomi/generated/graphql'
import type { ExtensionSource, ExtensionSourcePreferenceModel, ExtensionSourcePreferenceType, UpdatePreferenceParams } from '../../../extension.domain'
import { ExtensionModel } from '../../../extension.domain'

// ── Extension / Source nodes ──────────────────────────────────────────────────

export interface ExtensionNode {
  pkgName: string
  name: string
  lang: string
  iconUrl: string
  isNsfw: boolean
  isInstalled: boolean
  hasUpdate: boolean
  versionName: string
}

export function extensionToDomain(node: ExtensionNode): ExtensionModel {
  return new ExtensionModel({
    pkgName: node.pkgName,
    name: node.name,
    lang: node.lang,
    iconUrl: node.iconUrl || undefined,
    isNsfw: node.isNsfw,
    isInstalled: node.isInstalled,
    hasUpdate: node.hasUpdate,
    versionName: node.versionName,
  })
}

export interface SourceNode {
  id: string
  name: string
  lang: string
  isNsfw: boolean
  isConfigurable: boolean
}

export function sourceToDomain(node: SourceNode): ExtensionSource {
  return {
    id: node.id,
    name: node.name,
    lang: node.lang,
    isNsfw: node.isNsfw,
    isConfigurable: node.isConfigurable,
  }
}

// ── Preference union — generated discriminated union ──────────────────────────

// Derive the element type directly from the generated query type so TypeScript
// narrows each branch to the aliased fields without any unsafe cast.
type PreferencesArray = GetSourcePreferencesQuery['source']['preferences']
export type PreferenceNode = PreferencesArray[number]

export function preferenceToDomain(node: PreferenceNode, index: number): ExtensionSourcePreferenceModel {
  const base = {
    position: index,
    key: node.key ?? undefined,
    title: node.title ?? undefined,
    summary: node.summary ?? undefined,
    visible: node.visible,
  }

  switch (node.__typename) {
    case 'SwitchPreference':
      return {
        ...base,
        type: 'switch',
        booleanValue: node.currentValueBool ?? undefined,
        booleanDefault: node.defaultBool,
      }
    case 'CheckBoxPreference':
      return {
        ...base,
        type: 'checkbox',
        booleanValue: node.currentValueBool ?? undefined,
        booleanDefault: node.defaultBool,
      }
    case 'EditTextPreference':
      return {
        ...base,
        type: 'editText',
        textValue: node.currentValueStr ?? undefined,
        textDefault: node.defaultStr ?? undefined,
        dialogTitle: node.dialogTitle ?? undefined,
        dialogMessage: node.dialogMessage ?? undefined,
      }
    case 'ListPreference':
      return {
        ...base,
        type: 'list',
        textValue: node.currentValueStr ?? undefined,
        textDefault: node.defaultStr ?? undefined,
        entries: node.entries,
        entryValues: node.entryValues,
      }
    case 'MultiSelectListPreference':
      return {
        ...base,
        type: 'multiSelect',
        multiValue: node.currentValueList ?? undefined,
        multiDefault: node.defaultList ?? undefined,
        entries: node.entries,
        entryValues: node.entryValues,
      }
    default: {
      const exhaustive: never = node

      throw new Error(`Unknown preference typename: ${(exhaustive as { __typename: string }).__typename}`)
    }
  }
}

// ── Change input builder ──────────────────────────────────────────────────────

export interface ChangeInput {
  position: number
  switchState?: boolean
  checkBoxState?: boolean
  editTextState?: string
  listState?: string
  multiSelectState?: string[]
}

export function toChangeInput(type: ExtensionSourcePreferenceType, p: UpdatePreferenceParams): ChangeInput {
  switch (type) {
    case 'switch':
      return { position: p.position, switchState: p.booleanValue }
    case 'checkbox':
      return { position: p.position, checkBoxState: p.booleanValue }
    case 'editText':
      return { position: p.position, editTextState: p.textValue }
    case 'list':
      return { position: p.position, listState: p.textValue }
    case 'multiSelect':
      return { position: p.position, multiSelectState: p.multiValue }
  }
}
