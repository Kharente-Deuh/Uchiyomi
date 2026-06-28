// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { ExtensionSettings, ExtensionSourceRepository, SuwayomiExtensionsPort } from '../../extension.domain'
import { mergeExtensionSettings } from '../../extension.domain'

export interface GetExtensionSettingsUseCaseOpts { pkgName: string }

export class GetExtensionSettingsUseCase implements IUseCase<GetExtensionSettingsUseCaseOpts, ExtensionSettings> {
  constructor(
    private readonly sources: ExtensionSourceRepository,
    private readonly suwayomi: SuwayomiExtensionsPort,
  ) {}

  async execute({ pkgName }: GetExtensionSettingsUseCaseOpts): Promise<ExtensionSettings> {
    const configurable = await this.sources.findMany({ pkgName, isConfigurable: true })
    const withPrefs = await Promise.all(configurable.map(async source => ({
      id: source.id,
      name: source.name,
      lang: source.lang,
      preferences: await this.suwayomi.listSourcePreferences(source.id),
    })))

    return mergeExtensionSettings(withPrefs)
  }
}
