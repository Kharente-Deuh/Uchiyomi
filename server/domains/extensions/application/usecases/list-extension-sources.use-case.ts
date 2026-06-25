// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourceRepository, StoredExtensionSource } from '../../extension.domain'

export interface ListExtensionSourcesUseCaseOpts {
  pkgName: string
  isAdmin: boolean
  viewerCanSeeNsfw: boolean
}

export class ListExtensionSourcesUseCase implements IUseCase<ListExtensionSourcesUseCaseOpts, StoredExtensionSource[]> {
  constructor(private readonly sources: ExtensionSourceRepository) {}

  async execute(opts: ListExtensionSourcesUseCaseOpts): Promise<StoredExtensionSource[]> {
    const all = await this.sources.listByPkg(opts.pkgName)
    if (opts.isAdmin) {
      return all
    }

    return all
      .filter(s => s.isEnabled)
      .filter(s => opts.viewerCanSeeNsfw || !s.isNsfw)
  }
}
