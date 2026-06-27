// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { ExtensionSourceRepository, StoredExtensionSource } from '../../extension.domain'

export interface ListExtensionSourcesUseCaseOpts {
  pkgName: string
  isAdmin: boolean
  canSeeNsfw: boolean
}

export class ListExtensionSourcesUseCase implements IUseCase<ListExtensionSourcesUseCaseOpts, StoredExtensionSource[]> {
  constructor(private readonly sources: ExtensionSourceRepository) {}

  execute(opts: ListExtensionSourcesUseCaseOpts): Promise<StoredExtensionSource[]> {
    return this.sources.findMany({
      pkgName: opts.pkgName,
      ...(!opts.isAdmin && { isEnabled: true }),
      ...(!opts.canSeeNsfw && { isNsfw: false }),
    })
  }
}
