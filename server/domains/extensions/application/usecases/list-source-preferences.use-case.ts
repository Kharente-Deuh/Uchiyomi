// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourcePreferenceModel, SuwayomiExtensionsPort } from '../../extension.domain'

export interface ListSourcePreferencesUseCaseOpts { sourceId: string }

export class ListSourcePreferencesUseCase implements IUseCase<ListSourcePreferencesUseCaseOpts, ExtensionSourcePreferenceModel[]> {
  constructor(private readonly suwayomi: SuwayomiExtensionsPort) {}

  execute(opts: ListSourcePreferencesUseCaseOpts): Promise<ExtensionSourcePreferenceModel[]> {
    return this.suwayomi.listSourcePreferences(opts.sourceId)
  }
}
