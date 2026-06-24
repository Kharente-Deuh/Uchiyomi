// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { SettingsPort } from '../../suwayomi-settings.domain'
import { DESIRED_SETTINGS, reconcile } from '../../suwayomi-settings.domain'

export interface ReconcileSettingsUseCaseResult {
  changed: boolean
  changes: string[]
}

export class ReconcileSettingsUseCase implements IUseCase<void, ReconcileSettingsUseCaseResult> {
  constructor(private readonly settings: SettingsPort) {}

  async execute(): Promise<ReconcileSettingsUseCaseResult> {
    const current = await this.settings.read()
    const { patch, changes } = reconcile(current, DESIRED_SETTINGS)
    if (changes.length > 0) {
      await this.settings.apply(patch)
    }

    return { changed: changes.length > 0, changes }
  }
}
