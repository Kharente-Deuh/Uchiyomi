// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { ExtensionModel, ExtensionSourceRepository, SuwayomiExtensionsPort } from '../../extension.domain'

export interface UpdateExtensionUseCaseOpts {
  pkgName: string
}

export class UpdateExtensionUseCase implements IUseCase<UpdateExtensionUseCaseOpts, ExtensionModel> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly sources: ExtensionSourceRepository,
  ) {}

  async execute(opts: UpdateExtensionUseCaseOpts): Promise<ExtensionModel> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    // Updating only makes sense for an installed extension; the overlay row
    // already exists, so there is nothing to create first.
    if (!meta.isInstalled) {
      throw new Error(`Extension not installed: ${opts.pkgName}`)
    }

    await this.suwayomi.update(opts.pkgName)
    // A new version can add or drop sources, so re-sync the source overlay
    // (syncForExtension preserves each source's isEnabled flag).
    await this.sources.syncForExtension(opts.pkgName, await this.suwayomi.listSources(opts.pkgName))

    // Return the post-update state so callers don't need a refetch: the new
    // version reports hasUpdate=false and a bumped versionName. Fall back to the
    // pre-update meta if the refetch comes back empty.
    return await this.suwayomi.getExtension(opts.pkgName) ?? meta
  }
}
