// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourcePreferenceModel, SuwayomiExtensionsPort, UpdatePreferenceParams } from '../../extension.domain'

export class UpdateSourcePreferenceUseCase implements IUseCase<UpdatePreferenceParams, ExtensionSourcePreferenceModel[]> {
  constructor(private readonly suwayomi: SuwayomiExtensionsPort) {}

  execute(opts: UpdatePreferenceParams): Promise<ExtensionSourcePreferenceModel[]> {
    return this.suwayomi.updateSourcePreference(opts)
  }
}
