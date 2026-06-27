// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { ExtensionErrorLogEntry, ExtensionHealthRow, ExtensionsOverlayRepository } from '../../extension.domain'

export interface GetExtensionHealthUseCaseOpts { pkgName: string }
export interface GetExtensionHealthUseCaseResult { health: ExtensionHealthRow, log: ExtensionErrorLogEntry[] }

export class GetExtensionHealthUseCase implements IUseCase<GetExtensionHealthUseCaseOpts, GetExtensionHealthUseCaseResult | undefined> {
  constructor(private readonly overlay: ExtensionsOverlayRepository) {}

  async execute(opts: GetExtensionHealthUseCaseOpts): Promise<GetExtensionHealthUseCaseResult | undefined> {
    const health = await this.overlay.findHealth(opts.pkgName)
    if (!health) {
      return undefined
    }

    const log = await this.overlay.listErrorLog(opts.pkgName)

    return { health, log }
  }
}
