// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionSourceRepository, ExtensionsOverlayRepository, SuwayomiExtensionsPort } from '../../extension.domain'

export interface InstallExtensionUseCaseOpts {
  pkgName: string
  actorId: string
}

export class InstallExtensionUseCase implements IUseCase<InstallExtensionUseCaseOpts, void> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
    private readonly sources: ExtensionSourceRepository,
  ) {}

  async execute(opts: InstallExtensionUseCaseOpts): Promise<void> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    // Create the overlay row up front so health/error-log have a parent to attach to.
    await this.overlay.upsertInstalled({
      pkgName: meta.pkgName,
      name: meta.name,
      lang: meta.lang,
      iconUrl: meta.iconUrl,
      isNsfw: meta.isNsfw,
      installedByUserId: opts.actorId,
    })
    try {
      await this.suwayomi.install(opts.pkgName)
      await this.overlay.recordSuccess(opts.pkgName)
      await this.sources.syncForExtension(opts.pkgName, await this.suwayomi.listSources(opts.pkgName))
    } catch (err) {
      await this.overlay.recordFailure({ pkgName: opts.pkgName, message: (err as Error).message, context: 'install' })

      throw err
    }
  }
}
