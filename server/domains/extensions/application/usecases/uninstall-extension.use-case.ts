// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionsOverlayRepository, ListedExtension, SuwayomiExtensionsPort } from '../../extension.domain'
import { toListedExtension } from '../../extension.domain'

export interface UninstallExtensionUseCaseOpts {
  pkgName: string
}

export class UninstallExtensionUseCase implements IUseCase<UninstallExtensionUseCaseOpts, ListedExtension> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
  ) {}

  async execute(opts: UninstallExtensionUseCaseOpts): Promise<ListedExtension> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    await this.suwayomi.uninstall(opts.pkgName)
    await this.overlay.deleteByPkgName(opts.pkgName)

    // The extension is now uninstalled and its overlay health row is gone, so
    // health is omitted (toListedExtension yields isHealthy: undefined anyway).
    return toListedExtension({ ...meta, isInstalled: false })
  }
}
