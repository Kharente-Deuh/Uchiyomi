// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionsOverlayRepository, SuwayomiExtensionsPort } from '../../extension.domain'

export interface UninstallExtensionUseCaseOpts {
  pkgName: string
}

export class UninstallExtensionUseCase implements IUseCase<UninstallExtensionUseCaseOpts, void> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
  ) {}

  async execute(opts: UninstallExtensionUseCaseOpts): Promise<void> {
    await this.suwayomi.uninstall(opts.pkgName)
    await this.overlay.deleteByPkgName(opts.pkgName)
  }
}
