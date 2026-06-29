// SPDX-License-Identifier: AGPL-3.0-or-later

import type { IUseCase, Page } from '~~/server/shared'
import type { ExtensionModel, ListExtensionsFilters, SuwayomiExtensionsPort } from '../../extension.domain'

export interface ListExtensionsUseCaseFilters {
  pkgName?: string
  search?: string
  isInstalled?: boolean
  hasUpdate?: boolean
  nsfw?: boolean
}

export interface ListExtensionsUseCaseOpts {
  isAdmin: boolean
  viewerCanSeeNsfw: boolean
  page: number
  pageSize: number
  filters?: ListExtensionsUseCaseFilters
}

export class ListExtensionsUseCase implements IUseCase<ListExtensionsUseCaseOpts, Page<ExtensionModel>> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
  ) {}

  async execute({ isAdmin, viewerCanSeeNsfw, page, pageSize, filters }: ListExtensionsUseCaseOpts): Promise<Page<ExtensionModel>> {
    // Visibility/NSFW are server-enforced and override client-supplied values:
    // non-admins only see installed extensions; viewers without NSFW permission
    // never see NSFW regardless of the client `nsfw` flag.
    const effectiveFilters: ListExtensionsFilters = {
      pkgName: filters?.pkgName,
      search: filters?.search,
      hasUpdate: filters?.hasUpdate,
      isInstalled: isAdmin ? filters?.isInstalled : true,
      isNsfw: viewerCanSeeNsfw ? filters?.nsfw : false,
    }

    return this.suwayomi.listExtensions({ page, pageSize, filters: effectiveFilters })
  }
}
