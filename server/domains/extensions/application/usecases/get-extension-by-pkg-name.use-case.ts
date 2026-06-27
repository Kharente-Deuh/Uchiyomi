// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '~~/server/shared'
import type { ExtensionModel, SuwayomiExtensionsPort } from '../../extension.domain'

export interface GetExtensionByPkgNameUseCaseOpts { pkgName: string }

export class GetExtensionByPkgNameUseCase implements IUseCase<GetExtensionByPkgNameUseCaseOpts, ExtensionModel | undefined> {
  constructor(private readonly suwayomi: SuwayomiExtensionsPort) {}

  async execute(opts: GetExtensionByPkgNameUseCaseOpts): Promise<ExtensionModel | undefined> {
    const extension = await this.suwayomi.getExtension(opts.pkgName)
    if (!extension) {
      return
    }

    return extension
  }
}
