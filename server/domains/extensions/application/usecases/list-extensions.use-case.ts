// SPDX-License-Identifier: AGPL-3.0-or-later
import type { IUseCase } from '../../../../shared/use-case'
import type { ExtensionsOverlayRepository, ListedExtension, ListExtensionsFilters, SuwayomiExtensionsPort } from '../../extension.domain'
import { toListedExtension } from '../../extension.domain'

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

export interface ListExtensionsUseCaseResult {
  items: ListedExtension[]
  page: number
  pageSize: number
  totalCount: number
}

export class ListExtensionsUseCase implements IUseCase<ListExtensionsUseCaseOpts, ListExtensionsUseCaseResult> {
  constructor(
    private readonly suwayomi: SuwayomiExtensionsPort,
    private readonly overlay: ExtensionsOverlayRepository,
  ) {}

  async execute({ isAdmin, viewerCanSeeNsfw, page, pageSize, filters }: ListExtensionsUseCaseOpts): Promise<ListExtensionsUseCaseResult> {
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

    const pageResult = await this.suwayomi.listExtensions({ page, pageSize, filters: effectiveFilters })
    const healthRows = await this.overlay.listHealthByPkgNames(pageResult.items.map(e => e.pkgName))
    const healthByPkg = new Map(healthRows.map(h => [h.pkgName, h]))
    const items = pageResult.items.map(e => toListedExtension(e, healthByPkg.get(e.pkgName)))

    return { items, page, pageSize, totalCount: pageResult.totalCount }
  }
}
