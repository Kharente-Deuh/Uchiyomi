// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ManagedSettings, SettingsPort } from '../../server/domains/suwayomi-settings/suwayomi-settings.domain'
import { describe, expect, it, vi } from 'vitest'
import * as ReconcileSettings from '../../server/domains/suwayomi-settings/application/usecases/reconcile-settings.use-case'

const KEIYOUSHI = 'https://github.com/keiyoushi/extensions/tree/repo'

function fakePort(current: ManagedSettings): SettingsPort & { applied: unknown[] } {
  const applied: unknown[] = []

  return {
    applied,
    read: async () => current,
    apply: async (patch) => {
      applied.push(patch)
    },
  }
}

describe('reconcileSettings.UseCase', () => {
  it('applies a patch and reports the changes when settings drift', async () => {
    const port = fakePort({ autoDownloadNewChapters: true, extensionRepos: null })
    const report = await new ReconcileSettings.ReconcileSettingsUseCase(port).execute()
    expect(report.changed).toBe(true)
    expect(report.changes).toHaveLength(2)
    expect(port.applied).toEqual([{ autoDownloadNewChapters: false, extensionRepos: [KEIYOUSHI] }])
  })

  it('does not call apply when already in desired state', async () => {
    const port = fakePort({ autoDownloadNewChapters: false, extensionRepos: [KEIYOUSHI] })
    const applySpy = vi.spyOn(port, 'apply')
    const report = await new ReconcileSettings.ReconcileSettingsUseCase(port).execute()
    expect(report).toEqual({ changed: false, changes: [] })
    expect(applySpy).not.toHaveBeenCalled()
  })
})
