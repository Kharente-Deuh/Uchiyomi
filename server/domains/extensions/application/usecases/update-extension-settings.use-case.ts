// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { ExtensionSettings, ExtensionSourceRepository, SuwayomiExtensionsPort } from '../../extension.domain'
import { computeSourceChanges, mergeExtensionSettings } from '../../extension.domain'

export interface UpdateExtensionSettingsUseCaseOpts {
  pkgName: string
  settings: ExtensionSettings
}

export class UpdateExtensionSettingsUseCase implements IUseCase<UpdateExtensionSettingsUseCaseOpts, ExtensionSettings> {
  constructor(
    private readonly sources: ExtensionSourceRepository,
    private readonly suwayomi: SuwayomiExtensionsPort,
  ) {}

  async execute({ pkgName, settings }: UpdateExtensionSettingsUseCaseOpts): Promise<ExtensionSettings> {
    const configurable = await this.sources.findMany({ pkgName, isConfigurable: true })

    // Apply per source; one failed source must not abort the rest.
    await Promise.all(configurable.map(async (source) => {
      try {
        const current = await this.suwayomi.listSourcePreferences(source.id)
        const perSource = settings.sources.find(s => s.id === source.id)
        const changes = computeSourceChanges(current, perSource, settings.common)
        if (changes.length > 0) {
          await this.suwayomi.updateSourcePreferences(source.id, changes)
        }
      } catch (error) {
        console.warn(`[extension-settings] failed to apply settings to source ${source.id}:`, error)
      }
    }))

    // Re-read the persisted truth and re-merge.
    const withPrefs = await Promise.all(configurable.map(async source => ({
      id: source.id,
      name: source.name,
      lang: source.lang,
      preferences: await this.suwayomi.listSourcePreferences(source.id),
    })))

    return mergeExtensionSettings(withPrefs)
  }
}
