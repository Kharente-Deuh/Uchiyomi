// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as Extension`.

export type ExtensionHealth = 'OK' | 'ERROR'

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

export interface ExtensionHealthRow {
  pkgName: string
  health: ExtensionHealth
  consecutiveFailures: number
  lastErrorAt?: Date
  lastErrorMessage?: string
  installedByUserId?: string
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

export interface Page<T> {
  items: T[]
  totalCount: number
}

export interface ListedExtension extends ExtensionModel {
  isHealthy?: boolean
}

// Annotate a Suwayomi extension with overlay-derived health. Health only applies
// to installed extensions (undefined otherwise); a missing health row is treated
// as healthy. Shared by the list / install / uninstall / update use cases so the
// isHealthy rule lives in one place.
export function toListedExtension(ext: ExtensionModel, health?: ExtensionHealthRow): ListedExtension {
  return {
    ...ext,
    isHealthy: ext.isInstalled ? (health ? health.health === 'OK' : true) : undefined,
  }
}

export interface ExtensionSource {
  id: string
  name: string
  lang: string
  isNsfw: boolean
  isConfigurable: boolean
}

export type ExtensionSourcePreferenceType = 'switch' | 'checkbox' | 'editText' | 'list' | 'multiSelect'

export interface ExtensionSourcePreferenceModel {
  position: number
  type: ExtensionSourcePreferenceType
  key?: string
  title?: string
  summary?: string
  visible: boolean
  // value-bearing fields (only the relevant ones are set per type):
  booleanValue?: boolean // switch | checkbox (current)
  booleanDefault?: boolean
  textValue?: string // editText | list (current selected entryValue)
  textDefault?: string
  multiValue?: string[] // multiSelect
  multiDefault?: string[]
  entries?: string[] // list | multiSelect (labels)
  entryValues?: string[] // list | multiSelect (values)
  dialogTitle?: string
  dialogMessage?: string
}

export interface ExtensionErrorLogEntry {
  occurredAt: Date
  message: string
  context?: string
}

export interface StoredExtensionSource extends ExtensionSource {
  pkgName: string
  isEnabled: boolean
}

// ── Ports ──────────────────────────────────────────────────────────────────

export interface UpdatePreferenceParams {
  sourceId: string
  position: number
  // exactly one provided, matching the preference type at `position`:
  booleanValue?: boolean
  textValue?: string
  multiValue?: string[]
}

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
  updateSourcePreference: (p: UpdatePreferenceParams) => Promise<ExtensionSourcePreferenceModel[]>
}

export interface UpsertInstalledExtensionParams {
  pkgName: string
  name: string
  lang: string
  iconUrl?: string
  isNsfw: boolean
  installedByUserId: string
}

export interface RecordExtensionFailureParams {
  pkgName: string
  message: string
  context?: string
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

// Overlay (health + install trace) in PostgreSQL.
export interface ExtensionsOverlayRepository {
  upsertInstalled: (p: UpsertInstalledExtensionParams) => Promise<void>
  deleteByPkgName: (pkgName: string) => Promise<void>
  listHealthByPkgNames: (pkgNames: string[]) => Promise<ExtensionHealthRow[]>
  findHealth: (pkgName: string) => Promise<ExtensionHealthRow | undefined>
  recordFailure: (p: RecordExtensionFailureParams) => Promise<void>
  recordSuccess: (pkgName: string) => Promise<void>
  listErrorLog: (pkgName: string) => Promise<ExtensionErrorLogEntry[]>
}
