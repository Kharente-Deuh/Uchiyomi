// SPDX-License-Identifier: AGPL-3.0-or-later

// Flat ESM module, consumed via `import * as Extension`.

import type { Page } from '~~/server/shared'

export class ExtensionModel {
  declare pkgName: string
  declare name: string
  declare lang: string
  declare iconUrl?: string
  declare isNsfw: boolean
  declare isInstalled: boolean
  declare hasUpdate: boolean
  declare versionName: string

  constructor(data: ExtensionModel) {
    Object.assign<ExtensionModel, ExtensionModel>(this, data)
  }
}

export interface ListExtensionsFilters {
  pkgName?: string
  search?: string
  isInstalled?: boolean
  hasUpdate?: boolean
  isNsfw?: boolean
}

export interface ListExtensionsQuery {
  filters?: ListExtensionsFilters
  page: number
  pageSize: number
}

export interface ExtensionSource {
  id: string
  name: string
  lang: string
  isNsfw: boolean
  isConfigurable: boolean
  supportsLatest: boolean
}

export type ExtensionSourcePreferenceType = 'switch' | 'checkbox' | 'editText' | 'list' | 'multiSelect'

interface BasePreference {
  position: number
  key?: string
  title?: string
  summary?: string
  visible: boolean
}

interface SwitchPreference extends BasePreference { type: 'switch', booleanValue?: boolean, booleanDefault: boolean }
interface CheckboxPreference extends BasePreference { type: 'checkbox', booleanValue?: boolean, booleanDefault: boolean }
interface EditTextPreference extends BasePreference { type: 'editText', textValue?: string, textDefault?: string, dialogTitle?: string, dialogMessage?: string }
interface ListPreference extends BasePreference { type: 'list', textValue?: string, textDefault?: string, entries: string[], entryValues: string[] }
interface MultiSelectPreference extends BasePreference { type: 'multiSelect', multiValue?: string[], multiDefault?: string[], entries: string[], entryValues: string[] }

// Keep the exported name — it is referenced widely. Suwayomi-guaranteed fields
// (booleanDefault, entries, entryValues) are now required; currentValue fields
// stay optional (nullable in the Suwayomi schema).
export type ExtensionSourcePreferenceModel
  = | SwitchPreference
    | CheckboxPreference
    | EditTextPreference
    | ListPreference
    | MultiSelectPreference

export interface StoredExtensionSource extends ExtensionSource {
  pkgName: string
  isEnabled: boolean
}

// ── Ports ──────────────────────────────────────────────────────────────────

// Talks to Suwayomi (source of truth).
export interface SuwayomiExtensionsPort {
  listExtensions: (query: ListExtensionsQuery) => Promise<Page<ExtensionModel>>
  listAll: () => Promise<ExtensionModel[]>
  getExtension: (pkgName: string) => Promise<ExtensionModel | undefined>
  install: (pkgName: string) => Promise<void>
  uninstall: (pkgName: string) => Promise<void>
  update: (pkgName: string) => Promise<void>
  listSources: (pkgName: string) => Promise<ExtensionSource[]>
  listSourcePreferences: (sourceId: string) => Promise<ExtensionSourcePreferenceModel[]>
  updateSourcePreferences: (sourceId: string, changes: SourcePreferenceChange[]) => Promise<ExtensionSourcePreferenceModel[]>
}

export interface UpsertInstalledExtensionParams {
  pkgName: string
  name: string
  lang: string
  iconUrl?: string
  isNsfw: boolean
  installedByUserId: string
}

export interface FindManySourcesParams {
  pkgName: string
  isEnabled?: boolean
  isConfigurable?: boolean
  isNsfw?: boolean
  lang?: string
}

export interface ExtensionSourceRepository {
  syncForExtension: (pkgName: string, sources: ExtensionSource[]) => Promise<void>
  findMany: (params: FindManySourcesParams) => Promise<StoredExtensionSource[]>
  findById: (id: string) => Promise<StoredExtensionSource | undefined>
  update: (id: string, data: Partial<Omit<StoredExtensionSource, 'id'>>) => Promise<StoredExtensionSource>
}

// Overlay (install trace) in PostgreSQL.
export interface ExtensionsOverlayRepository {
  upsertInstalled: (p: UpsertInstalledExtensionParams) => Promise<void>
  deleteByPkgName: (pkgName: string) => Promise<void>
}

// ── Settings aggregate helpers ────────────────────────────────────────────────

// Longest common suffix of all strings in `strs`, character by character.
function longestCommonSuffix(strs: string[]): string {
  if (strs.length === 0) {
    return ''
  }

  const reversed = strs.map(s => [...s].toReversed().join(''))
  let common = reversed[0]!
  for (let i = 1; i < reversed.length; i++) {
    let j = 0
    while (j < common.length && j < reversed[i]!.length && common[j] === reversed[i]![j]) {
      j++
    }

    common = common.slice(0, j)
  }

  return [...common].toReversed().join('')
}

// Derive the discriminator suffix for a set of preference keys.
// Returns the longest trailing separator-delimited token common to all keys,
// e.g. '_af' or '_es-la'. Returns '' when there is no such token.
function deriveSuffix(keys: string[]): string {
  if (keys.length < 2) {
    return ''
  }

  const lcs = longestCommonSuffix(keys)
  const sepIdx = lcs.search(/[_-]/)
  if (sepIdx === -1) {
    return ''
  }

  return lcs.slice(sepIdx)
}

// Strip the discriminator suffix from a single key, if it ends with it.
function stemOf(key: string, suffix: string): string {
  if (suffix !== '' && key.endsWith(suffix)) {
    return key.slice(0, key.length - suffix.length)
  }

  return key
}

// Build a Map<stem, pref> for the keyed prefs of a single source.
// Includes a collision guard: if two distinct keyed prefs map to the same stem
// after stripping the computed suffix, fall back to raw keys (suffix = '').
function buildStemMap(
  prefs: ExtensionSourcePreferenceModel[],
): Map<string, ExtensionSourcePreferenceModel> {
  const keyed = prefs.filter(p => p.key !== undefined)
  const suffix = deriveSuffix(keyed.map(p => p.key!))

  const stemMap = new Map<string, ExtensionSourcePreferenceModel>()
  let collision = false

  for (const p of keyed) {
    const stem = stemOf(p.key!, suffix)
    if (stemMap.has(stem)) {
      collision = true
      break
    }

    stemMap.set(stem, p)
  }

  if (!collision) {
    return stemMap
  }

  // Collision detected: fall back to raw keys (always unique per source).
  const fallback = new Map<string, ExtensionSourcePreferenceModel>()
  for (const p of keyed) {
    fallback.set(p.key!, p)
  }

  return fallback
}

// ── Settings aggregate ("1 for all") ─────────────────────────────────────────

export interface ExtensionSettingsSource {
  id: string
  name: string
  lang: string
  preferences: ExtensionSourcePreferenceModel[] // non-common only
}

export interface ExtensionSettings {
  common: ExtensionSourcePreferenceModel[]
  sources: ExtensionSettingsSource[]
}

export interface MergeableExtensionSettingsSource {
  id: string
  name: string
  lang: string
  preferences: ExtensionSourcePreferenceModel[]
}

// Pure merge: a stem is "common" when every source exposes it with the same type
// (and same entryValues for list/multiSelect). Keys are normalized to stems by
// stripping each source's own discriminator suffix (e.g. '_af', '_es-la').
// Keyless prefs are never common. The common entry's value + key are taken from the
// reference (first) source, with key replaced by the stem; divergent per-source
// values converge on the next PUT.
export function mergeExtensionSettings(sources: MergeableExtensionSettingsSource[]): ExtensionSettings {
  if (sources.length === 0) {
    return { common: [], sources: [] }
  }

  // Build per-source stem maps.
  const stemMaps = sources.map(s => buildStemMap(s.preferences))

  // For each stem, track count, types, and entryValues across sources.
  const stemCount = new Map<string, number>()
  const stemTypes = new Map<string, Set<ExtensionSourcePreferenceType>>()
  // Store serialized entryValues per stem to detect mismatches (list/multiSelect only).
  const stemEntryValues = new Map<string, string | null>()

  for (const stemMap of stemMaps) {
    for (const [stem, p] of stemMap) {
      stemCount.set(stem, (stemCount.get(stem) ?? 0) + 1)
      const types = stemTypes.get(stem) ?? new Set<ExtensionSourcePreferenceType>()
      types.add(p.type)
      stemTypes.set(stem, types)

      if (p.type === 'list' || p.type === 'multiSelect') {
        const serial = JSON.stringify(p.entryValues ?? null)
        if (!stemEntryValues.has(stem)) {
          stemEntryValues.set(stem, serial)
        } else if (stemEntryValues.get(stem) !== serial) {
          // Mark as conflicting with a sentinel that will never match a real serial.
          stemEntryValues.set(stem, '\0conflict')
        }
      }
    }
  }

  const total = sources.length
  const isCommonStem = (stem: string): boolean => {
    const types = stemTypes.get(stem)
    if (stemCount.get(stem) !== total || types === undefined || types.size !== 1) {
      return false
    }

    const [type] = types
    if (type === 'list' || type === 'multiSelect') {
      return stemEntryValues.get(stem) !== '\0conflict'
    }

    return true
  }

  // Safe: sources.length > 0 is guaranteed by the early return above.
  const refStemMap = stemMaps[0]!

  // Collect common stems in reference-source order.
  const commonStems = new Set<string>()
  for (const stem of refStemMap.keys()) {
    if (isCommonStem(stem)) {
      commonStems.add(stem)
    }
  }

  return {
    common: [...refStemMap.entries()]
      .filter(([stem]) => commonStems.has(stem))
      .map(([stem, p]) => ({ ...p, key: stem })),
    sources: sources.map((source, i) => {
      const stemMap = stemMaps[i]!
      // Per-source: any pref whose stem is NOT common (including keyless prefs).
      const commonRawKeys = new Set<string>()
      for (const [stem, p] of stemMap) {
        if (commonStems.has(stem)) {
          commonRawKeys.add(p.key!)
        }
      }

      return {
        id: source.id,
        name: source.name,
        lang: source.lang,
        preferences: source.preferences.filter(p => p.key === undefined || !commonRawKeys.has(p.key)),
      }
    }),
  }
}

interface BaseChange { position: number }
export type SourcePreferenceChange
  = | (BaseChange & { type: 'switch', booleanValue: boolean })
    | (BaseChange & { type: 'checkbox', booleanValue: boolean })
    | (BaseChange & { type: 'editText', textValue: string })
    | (BaseChange & { type: 'list', textValue: string })
    | (BaseChange & { type: 'multiSelect', multiValue: string[] })

// Build a typed SourcePreferenceChange for the given position, carrying the
// value field from the preference model. Returns null when the branch's current
// value is undefined — a read-model echo with no concrete value produces no
// write (safe refinement: previously such a pref could emit a change carrying
// `undefined`, which the adapter would send to Suwayomi as a no-op at best).
function toSourceChange(position: number, p: ExtensionSourcePreferenceModel): SourcePreferenceChange | null {
  switch (p.type) {
    case 'switch':
      if (p.booleanValue === undefined) {
        return null
      }

      return { position, type: 'switch', booleanValue: p.booleanValue }
    case 'checkbox':
      if (p.booleanValue === undefined) {
        return null
      }

      return { position, type: 'checkbox', booleanValue: p.booleanValue }
    case 'editText':
      if (p.textValue === undefined) {
        return null
      }

      return { position, type: 'editText', textValue: p.textValue }
    case 'list':
      if (p.textValue === undefined) {
        return null
      }

      return { position, type: 'list', textValue: p.textValue }
    case 'multiSelect':
      if (p.multiValue === undefined) {
        return null
      }

      return { position, type: 'multiSelect', multiValue: p.multiValue }
  }
}

// Pure: given a source's freshly-read current prefs, produce the changes to apply.
// Common prefs carry a stem as their `key`; they are resolved to the live source
// position via the stem map. Per-source prefs match by raw key when present, else
// by position (keyless).
export function computeSourceChanges(
  current: ExtensionSourcePreferenceModel[],
  source: ExtensionSettingsSource | undefined,
  common: ExtensionSourcePreferenceModel[],
): SourcePreferenceChange[] {
  // Raw key → pref (for per-source resolution).
  const byKey = new Map<string, ExtensionSourcePreferenceModel>()
  for (const p of current) {
    if (p.key !== undefined) {
      byKey.set(p.key, p)
    }
  }

  // Stem → pref (for common resolution; c.key is a stem).
  const stemMap = buildStemMap(current)

  const changes: SourcePreferenceChange[] = []

  for (const c of common) {
    if (c.key === undefined) {
      continue
    }

    // c.key is a stem — look it up in the stem map.
    const target = stemMap.get(c.key)
    if (target && target.type === c.type) {
      const change = toSourceChange(target.position, c)
      if (change !== null) {
        changes.push(change)
      }
    }
  }

  for (const p of source?.preferences ?? []) {
    if (p.key === undefined) {
      const change = toSourceChange(p.position, p)
      if (change !== null) {
        changes.push(change)
      }
    } else {
      const target = byKey.get(p.key)
      if (target && target.type === p.type) {
        const change = toSourceChange(target.position, p)
        if (change !== null) {
          changes.push(change)
        }
      }
    }
  }

  return changes
}
