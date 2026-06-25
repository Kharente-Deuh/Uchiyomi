// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourceRepository, ExtensionsOverlayRepository, ListedExtension, SuwayomiExtensionsPort } from '../../extension.domain'
import { toListedExtension } from '../../extension.domain'

export interface UpdateExtensionUseCaseOpts {
  pkgName: string
}

export class UpdateExtensionUseCase implements IUseCase<UpdateExtensionUseCaseOpts, ListedExtension> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
    private readonly sources: ExtensionSourceRepository,
  ) {}

  async execute(opts: UpdateExtensionUseCaseOpts): Promise<ListedExtension> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    // Updating only makes sense for an installed extension; the overlay row (and
    // its health/error-log) already exists, so there is nothing to create first.
    if (!meta.isInstalled) {
      throw new Error(`Extension not installed: ${opts.pkgName}`)
    }

    try {
      await this.suwayomi.update(opts.pkgName)
      await this.overlay.recordSuccess(opts.pkgName)
      // A new version can add or drop sources, so re-sync the source overlay
      // (syncForExtension preserves each source's isEnabled flag).
      await this.sources.syncForExtension(opts.pkgName, await this.suwayomi.listSources(opts.pkgName))
    } catch (err) {
      await this.overlay.recordFailure({ pkgName: opts.pkgName, message: (err as Error).message, context: 'update' })

      throw err
    }

    // Return the post-update state so callers don't need a refetch: the new
    // version reports hasUpdate=false and a bumped versionName. Fall back to the
    // pre-update meta if the refetch comes back empty.
    const updated = await this.suwayomi.getExtension(opts.pkgName) ?? meta
    const health = await this.overlay.findHealth(opts.pkgName)

    return toListedExtension(updated, health)
  }
}
