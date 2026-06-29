// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase } from '~~/server/shared'
import type { ExtensionModel, ExtensionsOverlayRepository, SuwayomiExtensionsPort } from '../../extension.domain'

export interface UninstallExtensionUseCaseOpts {
  pkgName: string
}

export class UninstallExtensionUseCase implements IUseCase<UninstallExtensionUseCaseOpts, ExtensionModel> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
  ) {}

  async execute(opts: UninstallExtensionUseCaseOpts): Promise<ExtensionModel> {
    const meta = await this.suwayomi.getExtension(opts.pkgName)
    if (!meta) {
      throw new Error(`Extension not found: ${opts.pkgName}`)
    }

    await this.suwayomi.uninstall(opts.pkgName)
    await this.overlay.deleteByPkgName(opts.pkgName)

    return { ...meta, isInstalled: false }
  }
}
