// SPDX-License-Identifier: AGPL-3.0-or-later

import type { ManagedSettings } from '../../server/domains/suwayomi-settings/suwayomi-settings.domain'
import { describe, expect, it } from 'vitest'
import { DESIRED_SETTINGS, reconcile } from '../../server/domains/suwayomi-settings/suwayomi-settings.domain'

const KEIYOUSHI = 'https://github.com/keiyoushi/extensions/tree/repo'

describe('reconcile', () => {
  it('enforces a differing value (ENFORCE policy)', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: true, extensionRepos: [KEIYOUSHI] }
    const { patch, changes } = reconcile(current, DESIRED_SETTINGS)
    expect(patch).toEqual({ autoDownloadNewChapters: false })
    expect(changes).toHaveLength(1)
  })

  it('leaves a matching value untouched (ENFORCE policy)', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: false, extensionRepos: [KEIYOUSHI] }
    expect(reconcile(current, DESIRED_SETTINGS)).toEqual({ patch: {}, changes: [] })
  })

  it('fills an empty list (SEED_IF_EMPTY policy)', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: false, extensionRepos: [] }
    const { patch } = reconcile(current, DESIRED_SETTINGS)
    expect(patch).toEqual({ extensionRepos: [KEIYOUSHI] })
  })

  it('fills a null list (SEED_IF_EMPTY policy)', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: false, extensionRepos: null }
    const { patch } = reconcile(current, DESIRED_SETTINGS)
    expect(patch).toEqual({ extensionRepos: [KEIYOUSHI] })
  })

  it('leaves a populated list untouched (SEED_IF_EMPTY policy)', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: false, extensionRepos: ['https://example.com/repo'] }
    expect(reconcile(current, DESIRED_SETTINGS)).toEqual({ patch: {}, changes: [] })
  })

  it('combines both policies into one minimal patch', () => {
    const current: ManagedSettings = { autoDownloadNewChapters: true, extensionRepos: null }
    const { patch, changes } = reconcile(current, DESIRED_SETTINGS)
    expect(patch).toEqual({ autoDownloadNewChapters: false, extensionRepos: [KEIYOUSHI] })
    expect(changes).toHaveLength(2)
  })
})
