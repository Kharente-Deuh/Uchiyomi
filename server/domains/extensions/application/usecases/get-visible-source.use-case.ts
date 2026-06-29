// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { ExtensionSourceRepository, StoredExtensionSource } from '../../extension.domain'

export interface GetVisibleSourceUseCaseOpts {
  pkgName: string
  sourceId: string
  isAdmin: boolean
  canSeeNsfw: boolean
}

// Resolve a source for browsing/searching, applying the same visibility policy as
// listExtensionSources. Returns undefined for every "not visible" reason (unknown,
// wrong extension, disabled-for-non-admin, NSFW-without-permission) so the route can
// answer a single 404 without revealing which case occurred.
export class GetVisibleSourceUseCase implements IUseCase<GetVisibleSourceUseCaseOpts, StoredExtensionSource | undefined> {
  constructor(private readonly sources: ExtensionSourceRepository) {}

  async execute(opts: GetVisibleSourceUseCaseOpts): Promise<StoredExtensionSource | undefined> {
    const source = await this.sources.findById(opts.sourceId)
    if (!source || source.pkgName !== opts.pkgName) {
      return
    }

    if (!opts.isAdmin && !source.isEnabled) {
      return
    }

    if (!opts.canSeeNsfw && source.isNsfw) {
      return
    }

    return source
  }
}
