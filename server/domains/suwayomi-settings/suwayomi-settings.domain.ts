// SPDX-License-Identifier: AGPL-3.0-or-later
// Flat ESM module, consumed via `import * as SuwayomiSettings` where useful.

export type SettingPolicy = 'ENFORCE' | 'SEED_IF_EMPTY'

// The subset of Suwayomi settings the reconciler manages. `null` means "unset".
export interface ManagedSettings {
  autoDownloadNewChapters: boolean | null
  extensionRepos: string[] | null
}

// Only the keys that must change.
export type SettingsPatch = Partial<ManagedSettings>

export interface DesiredSetting {
  key: keyof ManagedSettings
  policy: SettingPolicy
  value: boolean | string[]
}

export interface ReconcileResult {
  patch: SettingsPatch
  changes: string[]
}

// Talks to Suwayomi (source of truth) for the managed subset.
export interface SettingsPort {
  read: () => Promise<ManagedSettings>
  apply: (patch: SettingsPatch) => Promise<void>
}

// Code-defined policy. Seeds the curated allowlist the M6.2 admin panel will extend.
export const DESIRED_SETTINGS: DesiredSetting[] = [
  // ADR-0012 invariant: all downloads go through the overlay queue.
  { key: 'autoDownloadNewChapters', policy: 'ENFORCE', value: false },
  // Keiyoushi is the de-facto community repo; seed so the catalogue works out of the box.
  { key: 'extensionRepos', policy: 'SEED_IF_EMPTY', value: ['https://github.com/keiyoushi/extensions/tree/repo'] },
]

function isEmpty(value: ManagedSettings[keyof ManagedSettings]): boolean {
  if (value == null) {
    return true
  }

  if (Array.isArray(value)) {
    return value.length === 0
  }

  return false
}

function equals(a: unknown, b: unknown): boolean {
  if (Array.isArray(a) && Array.isArray(b)) {
    return a.length === b.length && a.every((x, i) => x === b[i])
  }

  return a === b
}

export function reconcile(current: ManagedSettings, desired: DesiredSetting[]): ReconcileResult {
  const patch: SettingsPatch = {}
  const changes: string[] = []

  for (const setting of desired) {
    const value = current[setting.key]
    if (setting.policy === 'ENFORCE') {
      if (!equals(value, setting.value)) {
        // TypeScript cannot narrow the indexed write through a union key; the cast is safe
        // because `setting.value` is the declared value for that key.
        ;(patch as Record<string, unknown>)[setting.key] = setting.value
        changes.push(`${setting.key}: ${JSON.stringify(value)} -> ${JSON.stringify(setting.value)} (ENFORCE)`)
      }
    } else if (isEmpty(value)) {
      ;(patch as Record<string, unknown>)[setting.key] = setting.value
      changes.push(`${setting.key}: seeded ${JSON.stringify(setting.value)} (was empty)`)
    }
  }

  return { patch, changes }
}
