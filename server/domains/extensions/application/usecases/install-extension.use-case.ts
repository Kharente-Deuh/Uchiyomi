// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { ExtensionModel, ExtensionSourceRepository, ExtensionsOverlayRepository, SuwayomiExtensionsPort } from '../../extension.domain'

export interface InstallExtensionUseCaseOpts {
  pkgName: string
  actorId: string
}

export class InstallExtensionUseCase implements IUseCase<InstallExtensionUseCaseOpts, ExtensionModel> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
    private readonly sources: ExtensionSourceRepository,
  ) {}

  async execute(opts: InstallExtensionUseCaseOpts): Promise<ExtensionModel> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    // Create the overlay row up front so the install trace has a parent to attach to.
    await this.overlay.upsertInstalled({
      pkgName: meta.pkgName,
      name: meta.name,
      lang: meta.lang,
      iconUrl: meta.iconUrl,
      isNsfw: meta.isNsfw,
      installedByUserId: opts.actorId,
    })

    await this.suwayomi.install(opts.pkgName)
    await this.sources.syncForExtension(opts.pkgName, await this.suwayomi.listSources(opts.pkgName))

    // Return the now-installed state so callers don't need a refetch.
    return { ...meta, isInstalled: true }
  }
}
