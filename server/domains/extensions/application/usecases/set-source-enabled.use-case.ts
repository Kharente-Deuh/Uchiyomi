// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourceRepository, StoredExtensionSource } from '../../extension.domain'

export interface SetSourceEnabledUseCaseOpts {
  sourceId: string
  isEnabled: boolean
}

export class SetSourceEnabledUseCase implements IUseCase<SetSourceEnabledUseCaseOpts, StoredExtensionSource> {
  constructor(private readonly sources: ExtensionSourceRepository) {}

  async execute(opts: SetSourceEnabledUseCaseOpts): Promise<StoredExtensionSource> {
    const existing = await this.sources.findById(opts.sourceId)
    if (!existing) {
      throw new Error('Source not found')
    }

    if (opts.isEnabled === existing.isEnabled) {
      return existing
    }

    return this.sources.update(opts.sourceId, { isEnabled: opts.isEnabled })
  }
}
